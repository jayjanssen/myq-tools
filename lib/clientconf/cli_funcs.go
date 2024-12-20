package clientconf

// CLI Funcs can only really be tested from a cli

import (
	"flag"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-multierror"
)

// Set the standard MySQL flags we expect
func SetMySQLFlags() {
	flag.StringVar(&defaultsFile, "defaults-file", "", "mysql defaults file")

	flag.StringVar(&userFlag, "user", "", "mysql user, defaults to your username")
	flag.StringVar(&userFlag, "u", "", "short for -user")

	flag.StringVar(&passwordFlag, "password", "", "mysql password")
	flag.StringVar(&passwordFlag, "p", "", "short for -password")

	flag.StringVar(&hostFlag, "host", "", "mysql host, defaults to 127.0.0.1")
	flag.StringVar(&hostFlag, "h", "", "short for -host")

	flag.StringVar(&portFlag, "port", "", "mysql port, defaults to 3306")
	flag.StringVar(&portFlag, "P", "", "short for -port")

	flag.StringVar(&socketFlag, "socket", "", "mysql socket")
	flag.StringVar(&socketFlag, "S", "", "short for -socket")

	flag.StringVar(&sslCertFlag, "ssl-cert", "", "mysql ssl cert")
	flag.StringVar(&sslKeyFlag, "ssl-key", "", "mysql ssl key")
	flag.StringVar(&sslCaFlag, "ssl-ca", "", "mysql ssl CA")

	flag.BoolVar(&enableCleartextPlugin, "enable-cleartext-plugin", false, "mysql enable cleartext plugin")
}

// Creates a [https://pkg.go.dev/github.com/go-sql-driver/mysql#Config]('Config') option from the go-sql-driver/mysql from three sources:
// 1. Default connection settings
// 2. Parsing .my.cnf files & co. to get anything set not passed by flag
// 3. Command line arguments for necessary config flags
// Later settings override earlier.  I.e., command line arguments override .my.cnf file settings.
func GenerateConfig() (*mysql.Config, error) {
	var errs *multierror.Error

	// construct a cnf that merges our three sources
	cnf := initCnf()
	err := appendFiles(cnf, getCnfFiles())
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	applyFlags(cnf)

	// Translate cnf to mysql.Config
	config, err := cnfToConfig(cnf)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	return config, errs.ErrorOrNil()
}
