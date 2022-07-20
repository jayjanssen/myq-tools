package clientconf

import (
	"fmt"
	"os"
	"os/user"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/ini.v1"
)

// Find and read .my.cnf files

// mysql cnf files with possible [client] sections per: https://dev.mysql.com/doc/refman/8.0/en/option-files.html
func getCnfFiles() []string {
	var files = []string{
		`/etc/my.cnf`,
		`/etc/mysql/my.cnf`,
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

// Initialize a cnf
func initCnf() *ini.File {
	opts := ini.LoadOptions{
		AllowBooleanKeys: true,
		Loose:            true,
	}
	cnf := ini.Empty(opts)

	// Set some basic defaults
	username := `root`
	if user, err := user.Current(); err == nil {
		username = user.Username
	}
	cnf.NewSection(`client`)
	cnf.Section(`client`).NewKey(`user`, username)
	cnf.Section(`client`).NewKey(`host`, `127.0.0.1`)
	cnf.Section(`client`).NewKey(`port`, `3306`)

	return cnf
}

// Parse a file or string into the given cnf
func appendCnf(cnf *ini.File, input interface{}) error {
	return cnf.Append(input)
}

// Append each of the given files to the cnf
func appendFiles(cnf *ini.File, files []string) error {
	var errs *multierror.Error

	for _, file := range files {
		err := appendCnf(cnf, file)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs.ErrorOrNil()
}

// Command line flags
var userFlag string
var passwordFlag string
var hostFlag string
var portFlag string
var socketFlag string

// SSL stuff TODO
// * ssl-cert
// * ssl-key
// * ssl-ca
// * ssl-cipher

// Apply the flag variables to the given cnf [client] section
func applyFlags(cnf *ini.File) {
	if userFlag != "" {
		cnf.Section(`client`).NewKey(`user`, userFlag)
	}
	if passwordFlag != "" {
		cnf.Section(`client`).NewKey(`password`, passwordFlag)
	}
	if hostFlag != "" {
		cnf.Section(`client`).NewKey(`host`, hostFlag)
	}
	if portFlag != "" {
		cnf.Section(`client`).NewKey(`port`, portFlag)
	}
	if socketFlag != "" {
		cnf.Section(`client`).NewKey(`socket`, socketFlag)
	}
}

// Translate cnf to mysql.Config
func cnfToConfig(cnf *ini.File) *mysql.Config {
	config := mysql.NewConfig()
	if !cnf.HasSection(`client`) {
		return config
	}

	// clientMap is all the resolved settings
	clientMap := cnf.Section(`client`).KeysHash()

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

	return config
}
