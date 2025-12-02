// Package dsndriver provides a MySQL driver that can hot swap the DSN.
// The driver is registered as "mysql-hotswap-dsn" and is a transparent, drop-in
// replacement for the real Go MySQL driver: github.com/go-sql-driver/mysql.
// To use this driver:
//
//   import dsndriver "github.com/go-mysql/hotswap-dsn-driver"
//
//   // Set hot swap callback function only once, at start
//   dsndriver.SetHotswapFunc(func(ctx context.Context, currentDSN string) (newDSN string) {
//       // User-provided code to load and return new DSN
//       // if it has changed, else return an empty string.
//       return "user:new-pass@tcp(127.0.0.1)/"
//   })
//
//   db, err := sql.Open("mysql-hotswap-dsn", "user:pass@tcp(127.0.0.1)/")
//
// Then use the db as normal. This driver only implement connection-related
// interface, and it only hot swaps the DSN by calling the hot sap function
// when a new connection returns MySQL error code 1045 (access denied).
// All other functionality is handled by the real MySQL driver directly.
//
// To use this driver, it is not necessary to import github.com/go-sql-driver/mysql.
// This package and github.com/go-sql-driver/mysql can be imported together if
// the latter is needed for its exported identifiers.
//
// The first connection to return MySQL error 1045 (access denied) calls the
// hotswap function and blocks other failed connections until it returns.
// Once the hotswap function returns, the first failed connection is retried
// with the new DSN. Once this returns (successful or not), it unblocks other
// failed/waiting connections which also retry (in parallel) with the new DSN.
// If the new DSN works, all connections return successfully to the caller and
// no errors are reported. If the new DSN does not work, the process is repeated.
// There is currently no TTL, backoff, or cooldown period between hotswaps.
package dsndriver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"sync"
	"sync/atomic"

	"github.com/go-sql-driver/mysql"
)

// Debug prints debug info using the Go log package, if true.
var Debug bool = false

// MySQLDriver implements driver.Driver and driver.DriverContext.
type MySQLDriver struct{}

func (d MySQLDriver) Open(dsn string) (driver.Conn, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	myc, err := newConnector(cfg)
	if err != nil {
		return nil, err
	}
	c := NewConnector(dsn, myc)
	return c.Connect(context.Background())
}

func (d MySQLDriver) OpenConnector(dsn string) (driver.Connector, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	myc, err := newConnector(cfg)
	if err != nil {
		return nil, err
	}
	return NewConnector(dsn, myc), nil
}

// --------------------------------------------------------------------------

// swapper defaults to a no-op func. User should call SetDSNHotswapFunc to set
// a real function to hot swap the DSN.
var swapper func(context.Context, string) string = nopSwapper

func nopSwapper(_ context.Context, _ string) string {
	return ""
}

// SetHotswapFunc sets the callback function to hot swap the DSN. This  must
// be set only once before any calls to sql.Open. It is not safe to set again or
// at any other time.
//
// The current DSN is passed to the callback function which should return the new
// DSN if and only if the DSN has changed. Any non-empty return string is used as
// the new DSN. If the DSN has not changed, or if there is an error, return an empty
// string.
//
// The callback function must respect the context and return an empty string if
// the context is canceled.
//
// The callback function is called serially. The first connection to return
// MySQL error code 1045 (access denied) will invoke the callback. While the
// callback is running, other connections that fail with MySQL error code 1045
// will wait on the first to hot swap the DSN. All connections respect the
// context. The callback is abandoned (its return value ignored) if the context
// is canceled while it is running.
func SetHotswapFunc(f func(ctx context.Context, currentDSN string) (newDSN string)) {
	swapper = f
}

// --------------------------------------------------------------------------

// newConnector returns a new mysql.Connector by default, but for testing
// we override to return a mockConnector so myc atomic.Value (below) doesn't
// panic on different data types.
var newConnector func(*mysql.Config) (driver.Connector, error) = mysql.NewConnector

// connector wraps a mysql.Connector. Both implement driver.Connector.
type connector struct {
	// myc stores the current mysql.Connector which makes real connections
	// to MySQL. This connector wraps conn (both implement drver.Connector).
	// When the DSN changes, we throw away the old mysql.Connector (let Go
	// garbage collect it) and store the new mysql.Connector with the new DSN.
	//
	// It's important to know that the driver.Connector is not stateful,
	// which is why we can hot swap it. In the source code for database/sql.go:
	//
	//   func OpenDB(c driver.Connector) *DB {
	//       ctx, cancel := context.WithCancel(context.Background())
	//       db := &DB{
	//           connector:    c,
	//
	// The *sql.DB stores only one driver.Connector (c). This code works because
	// c = &dsndriver.connector{} which does not change in the *sql.DB.
	// This allows us to hot swap the mysql.Connector without affecting the *sql.DB
	// or leaking connection pools. Connections in the *sql.DB can be from any
	// Connector becuase they're not tied to the Connector that creates them.
	myc atomic.Value

	// The mutex guarantees only 1 caller in Connect calls the hot swap func (swapper).
	// It's only checked when a connection gets MySQL error code 1045, so the
	// mutex is not in the fast path (i.e. when everything is ok). On MySQL error code 1045,
	// the first caller to lock and see that swapping = false is the "winner"
	// which calls the hot swap func. It also creates the ready chan to broadcast
	// to other callers who wait on the first. When the first is done, it closes
	// the chan and all callers try to connect again.
	//
	// Do not guard myc atomic.Value! It's atomic and safe for concurrent access.
	*sync.Mutex
	swapping bool
	ready    chan struct{}
	dsn      string
}

// NewConnector creates a new connector that wraps a mysql.Connector.
// Do not call this function; it is called by the driver.
func NewConnector(dsn string, myc driver.Connector) *connector {
	var m atomic.Value
	m.Store(myc)
	return &connector{
		myc:   m,
		Mutex: &sync.Mutex{},
		dsn:   dsn,
	}
}

func (h *connector) Connect(ctx context.Context) (driver.Conn, error) {
	// Call mysql.Connector to make the connection. When all is ok, this returns
	// a driver.Conn and we return early--no locking in this pkg.
	myc := h.myc.Load().(driver.Connector)
	conn, myerr := myc.Connect(ctx)
	if myerr == nil {
		return conn, nil // conn OK
	}

	// Connection failed. Return early if the error is not the only one we care
	// about: MySQL error code 1045 (access denied).
	if val, ok := myerr.(*mysql.MySQLError); !ok || val.Number != 1045 {
		return nil, myerr // conn fail but not "access denied"
	}

	// -----------------------------------------------------------------------
	// Hot swap DSN when conn fails with MySQL error code 1045 (access denied)
	// -----------------------------------------------------------------------

	// There can be many conn at this point (or perhaps just one unlucky conn).
	// First, lock the shared mutex and see if another conn is already swapping...
	h.Lock()
	if h.swapping {
		h.Unlock()
		// We're NOT the first failed conn, we're one of many that needs to wait
		// for the first conn to hot swap the DSN. The first conn will have already
		// created h.ready, and it'll close it when it's done swapping. So wait...
		select {
		case <-h.ready:
			// got new conn in time, retry conn
			return h.Connect(ctx)
		case <-ctx.Done():
		}
		return nil, ctx.Err()
	}

	// We're the winner: the very first failed conn to lock the mutex. Keep the
	// mutex while we set swapping=true and create the ready chan. These will
	// cause all other failed conns to wait in the block above.
	debug("hot swap begin")
	h.ready = make(chan struct{})
	h.swapping = true
	h.Unlock()

	defer func() {
		h.Lock()
		close(h.ready)     // unblock others waiting
		h.swapping = false // and swap again if necessary
		h.Unlock()
		debug("hot swap end")
	}()

	// We've released the lock but the following code is still serialized because
	// we set swapping=true which redirect other conns into the "if h.swapping {"
	// block.

	// Run the user-provided hot swap callback function in a goroutine so we can
	// wait here and abandon it if it takes too long. The done chan MUST be buffered
	// so we don't leak abandoned goroutines.
	done := make(chan string, 1)
	go func() {
		done <- swapper(ctx, h.dsn)
	}()

	// Waiting for the ^ hot swap callback func goroutine, or the context
	var newDSN string
	select {
	case newDSN = <-done:
		debug("new DSN: %s", newDSN)
	case <-ctx.Done():
		debug("timeout waiting for hot swap func (context canceled: %s)", ctx.Err())
		return nil, myerr
	}

	// No new DSN means either 1) the DSN didn't change or 2) the callback had
	// an error. Either way, there's nothing we can or should do here, so clean up
	// and return the original MySQL error. If we really did lose access to MySQL,
	// we'll probably keep hitting this code over and over until the hot swap func
	// returns a DSN that works.
	if newDSN == "" {
		return nil, myerr
	}

	// New DSN. Parse it and use it to create a new mysql.Connector. Return errors
	// here (not myerr, the original MySQL error) so the caller can see if they
	// returned a bad DSN.
	cfg, err := mysql.ParseDSN(newDSN)
	if err != nil {
		debug("mysql.ParseDSN error: %s", err)
		return nil, err
	}
	mycNew, err := newConnector(cfg)
	if err != nil {
		debug("mysql.NewConnector error: %s", err)
		return nil, err
	}
	h.myc.Store(mycNew) // hot swap the mysql.Connector with the new DSN
	h.dsn = newDSN      // store new DSN (don't need to guard)

	// Reconnect. DO NOT recurse (h.Connect(ctx)) because we lock and clean up
	// in the defer func ^, so if we recurse we'll dead lock on our self.
	return mycNew.Connect(ctx)
}

func (c *connector) Driver() driver.Driver {
	return &MySQLDriver{}
}

func debug(fmt string, args ...interface{}) {
	if !Debug {
		return
	}
	log.Printf("mysql-hotswap-dsn: "+fmt, args...)
}

func init() {
	sql.Register("mysql-hotswap-dsn", &MySQLDriver{})
}
