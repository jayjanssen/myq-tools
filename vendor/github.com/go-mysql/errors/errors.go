// Package errors provides functions and variables for handling common MySQL errors.
// Common errors are: connection-related, duplicate key, and read-only.
// These should be explicitly handled to allow retrying when possible
// (connection and read-only errors) and returning a better error to the caller
// (duplicate key).
//
// Handling these errors directly is not trivial. It requires low-level
// knowledge and experience with sql/database, github.com/go-sql-driver/mysql,
// and MySQL. See the implementation of Lost, for example. More importantly:
// these are implementation details which should not be leaked to the caller.
// The caller usually only needs to know if the connection was lost or MySQL
// is read-only (so it wait and retry), or if there was a duplicate key error
// (which may be an error to the user or something the caller expects and
// handles). This packages makes handling these errors easy.
//
// See https://dev.mysql.com/doc/refman/5.7/en/error-messages-server.html for
// a complete list of MySQL server errors. The vast majority are rarely
// encountered in typical, well-behaved systems.
//
// Error and CanRetry are the most important functions in this package. They are used like:
//
//   import my "github.com/go-mysql/errors"
//
//   SAVE_ITEM_LOOP:
//   for tryNo := 1; tryNo <= maxTries; tryNo++ {
//       if err := SaveItem(item); err != nil {
//           if ok, myerr := my.Error(err); ok { // MySQL error
//               if myerr == my.ErrDupeKey {
//                   return http.StatusConflict
//               }
//               if my.CanRetry(myerr) {
//                   time.Sleep(1 * time.Second)
//                   continue SAVE_ITEM_LOOP // retry
//               }
//           }
//           // Error not handled
//           return http.StatusInternalServerError
//       }
//       break SAVE_ITEM_LOOP // success
//   }
//
// The example above tries N-many times to save an item into a database.
// It handles MySQL errors explicitly. On ErrDupeKey it does not retry, it
// returns HTTP 409 (status conflict). Else if the error is transient, it
// waits 1 second and retries. All other errors cause an HTTP 500 (internal
// server error) return.
package errors

import (
	"database/sql/driver"
	"errors"
	"net"

	"github.com/go-sql-driver/mysql"
)

var (
	// ErrCannotConnect is returned when a connection cannot be made because
	// MySQL is unreachable for any reason. This is usually because MySQL is down,
	// but it could also indicate a network issue. If this error occurs frequently,
	// verify the network connection and MySQL address. If it occurs infrequently,
	// it could indicate a transient state that will recover automatically, e.g.
	// a failover.
	ErrCannotConnect = errors.New("cannot connect")

	// ErrConnLost is returned when the connection is lost, closed, or killed
	// for any reason. This could mean MySQL crashed, or a KILL command killed
	// connection. Lost implies the connection was previously connected.
	// If MySQL is up and ok, this could indicate a transient state that will
	// recover automatically. If not, ErrCannotConnect will probably be
	// returned next when the driver tries but fails to reestablish the connection.
	ErrConnLost = errors.New("connection lost")

	// ErrQueryKilled is returned when the KILL QUERY command is used. This only
	// kills the currently active query; the connection is still ok. Some tools
	// (like query snipers) use KILL QUERY to kill long-running queries. This
	// usually indicates that the program executing the query should try again.
	ErrQueryKilled = errors.New("query killed")

	// ErrReadOnly is returned when MySQL read-only is enabled.
	ErrReadOnly = errors.New("server is read-only")

	// ErrDupeKey is returned when a unique index prevents a value from being
	// inserted or updated. CanRetry returns false on this error.
	ErrDupeKey = errors.New("duplicate key value")
)

// Error returns an error in this package if possible. The boolean return
// value is true if the given error is any MySQL error. The error return value
// is an error in this package if the given error maps to one, else the given
// error is returned. If the given error is nil, Error returns false, nil.
func Error(err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	if Down(err) {
		return true, ErrCannotConnect
	}

	if Lost(err) {
		return true, ErrConnLost
	}

	errCode := MySQLErrorCode(err)
	if errCode == 0 {
		return false, err // not a MySQL error
	}
	switch errCode {
	case 1317: // ER_QUERY_INTERRUPTED
		return true, ErrQueryKilled
	case 1290, 1836: // ER_OPTION_PREVENTS_STATEMENT, ER_READ_ONLY_MODE
		return true, ErrReadOnly
	case 1062: // ER_DUP_ENTRY
		return true, ErrDupeKey
	}

	// A MySQL error, but not one we handle explicitly
	return true, err
}

// Down returns true if the error indicates MySQL cannot be reached for any
// reason. See ErrCannotConnect.
func Down(err error) bool {
	// Being unable to reach MySQL is a network issue, so we get a net.OpError.
	// If MySQL is reachable, then we'd get a mysql.* or driver.* error instead.
	_, ok := err.(*net.OpError)
	return ok
}

// Lost returns true if the error indicates the MySQL connection was lost.
// See ErrConnLost.
func Lost(err error) bool {
	// mysql.ErrInvalidConn is returned for sql.DB functions. driver.ErrBadConn
	// is returned for sql.Conn functions. These are the normal errors when
	// MySQL is lost.
	if err == mysql.ErrInvalidConn || err == driver.ErrBadConn {
		return true
	}

	// Server shutdown in progress is a special case: the conn will be lost
	// soon. The next call will most likely return in the block above ^.
	if errCode := MySQLErrorCode(err); errCode == 1053 { // ER_SERVER_SHUTDOWN
		return true
	}

	return false
}

// MySQLErrorCode returns the MySQL server error code for the error, or zero
// if the error is not a MySQL error. See https://dev.mysql.com/doc/refman/5.7/en/error-messages-server.html
func MySQLErrorCode(err error) uint16 {
	if val, ok := err.(*mysql.MySQLError); ok {
		return val.Number
	}
	return 0 // not a mysql error
}

// CanRetry returns true for every error in this package except ErrDupeKey.
// It returns false for all other errors, including nil.
func CanRetry(err error) bool {
	switch err {
	case ErrCannotConnect, ErrConnLost, ErrReadOnly, ErrQueryKilled:
		return true
	}
	return false
}
