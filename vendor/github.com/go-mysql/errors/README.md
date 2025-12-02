# MySQL Error Handling

[![Go Report Card](https://goreportcard.com/badge/github.com/go-mysql/errors)](https://goreportcard.com/report/github.com/go-mysql/errors) [![GoDoc](https://godoc.org/github.com/go-mysql/errors?status.svg)](https://godoc.org/github.com/go-mysql/errors)

`go-mysql/errors` provides functions and variables for handling common MySQL errors.

## Testing

Requires [MySQL Sandbox](https://github.com/datacharmer/mysql-sandbox). Install and export `MYSQL_SANDBOX_DIR` env var. For example: `MYSQL_SANDBOX_DIR=/Users/daniel/sandboxes/msb_5_7_21/` on a Mac. Tests take ~15s because the MySQL sandbox is restarted several times. Current test coverage: 100%.
