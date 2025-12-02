# MySQL Hotswap DSN Driver

[![Go Reference](https://pkg.go.dev/badge/github.com/go-mysql/hotswap-dsn-driver.svg)](https://pkg.go.dev/github.com/go-mysql/hotswap-dsn-driver)

This driver is a drop-in replacement for the real [Go MySQL driver](https://github.com/go-sql-driver/mysql) with one extra feature: it hotswaps the DSN on [MySQL error 1045 (access denied)](https://dev.mysql.com/doc/mysql-errors/5.7/en/server-error-reference.html#error_er_access_denied_error). This allows frequently rotating the MySQL password without app downtime or "access denied" errors.

Since the hotswap is handled at the driver-level, you don't have to handle the "access denied" error everywhere the `*sql.DB` is used. Instead, a single hotswap callback function is set:

```go
import dsndriver "github.com/go-mysql/hotswap-dsn-driver"

// Set hotswap callback function
dsndriver.SetHotswapFunc(func(ctx context.Context, currentDSN string) (newDSN string) {
    // Reload latest DSN and return.
    // Be sure to respect ctx, too.
    return "user:new_pass@tcp(127.0.0.1)/"
})

db, err := sql.Open("mysql-hotswap-dsn", "user:pass@tcp(127.0.0.1)/")

// Use db as usual
```

To use this driver, only two changes are required as shown above:

1. Set the hotswap callback func by calling `SetHotswapFunc`
2. Use driver name "mysql-hotswap-dsn" instead of "mysql": `sql.Open("mysql-hotswap-dsn", "<dsn>")`

When using this driver, you do _not_ need to import github.com/go-sql-driver/mysql (but you can if you need to).

The hotswap function is called _only_ for MySQL error 1045. Other errors are ignored. This driver only implements driver- and connector-related interfaces. None of the query-related interfaces are implemented, which means this driver never interferes with queries, transactions, etc.

See the [Go docs](https://pkg.go.dev/github.com/go-mysql/hotswap-dsn-driver) for additional information.
