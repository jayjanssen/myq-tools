package clientconf

import (
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-multierror"
)

func GenerateConfig() (*mysql.Config, error) {
	var errs *multierror.Error

	// Get cnf files
	files := getCnfFiles()
	cnf, err := loadFiles(files)
	if err != nil {
		// Do we care about cnf file errors?  Maybe just print a warning on startup?
		errs = multierror.Append(errs, err)
	}

	// Translate cnf to mysql.Config
	config := mysql.NewConfig()
	if client, err := cnf.GetSection(`client`); err != nil {
		clientMap := client.KeysHash()

		// Basic credentials
		if cnfval, ok := clientMap[`user`]; ok {
			config.User = cnfval
		}
		if cnfval, ok := clientMap[`password`]; ok {
			config.Passwd = cnfval
		}

		// Populate Net and Addr
		if cnfval, ok := clientMap[`socket`]; ok {
			config.Addr = cnfval
			config.Net = `unix`
		}

		// Host will override socket (I think that's how mysql does it?)
		if host, ok := clientMap[`host`]; ok {
			port, ok := clientMap[`port`]
			if !ok {
				port = `3306`
			}
			config.Addr = fmt.Sprintf("%s:%s", host, port)
			config.Net = `tcp`
		}

		// SSL stuff TODO
	}

	// Get command line arguments

	// Merge into mysql config

	return config, errs
}
