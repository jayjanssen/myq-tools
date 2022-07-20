package clientconf

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/ini.v1"
)

// Find and read .my.cnf files

// mysql cnf files with possible [client] sections per: https://dev.mysql.com/doc/refman/8.0/en/option-files.html
func getCnfFiles() []string {
	var files = []string{
		`/etc/my.cnf`,
		`/etc/mysql/my.cnf`,
		// `SYSCONFDIR/my.cnf`,

	}

	home, err := os.UserHomeDir()
	if err == nil {
		homedirFiles := []string{
			fmt.Sprintf(`%s/.my.cnf`, home),
			fmt.Sprintf(`%s/.mylogin.cnf`, home),
		}
		files = append(files, homedirFiles...)
	}

	return files
}

// Parse
func loadFiles(files []string) (*ini.File, error) {
	opts := ini.LoadOptions{
		AllowBooleanKeys: true,
	}
	cfg := ini.Empty(opts)
	var errs *multierror.Error

	for _, file := range files {
		err := cfg.Append(file)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return cfg, errs.ErrorOrNil()
}
