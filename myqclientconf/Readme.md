
# MySQL connection generator

Creates a [https://pkg.go.dev/github.com/go-sql-driver/mysql#Config]('Config') option from the go-sql-driver/mysql from two sources:

1. Command line arguments for necessary config flags
2. Parsing .my.cnf files & co. to get anything set not passed by flag

The command line arguments can easily be added to any cli using the `flag` go lib and any conflicting flags with the cli are raised as errors

Supported parameters:
* host
* port
* socket
* user
* password 
* ssl-cert
* ssl-key
* ssl-ca
* ssl-cipher
* ??
