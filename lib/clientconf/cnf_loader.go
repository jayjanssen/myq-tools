package clientconf

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
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
var sslCertFlag string
var sslKeyFlag string
var sslCaFlag string

// ssl cipher support TODO.  MySQL cipher names don't match go's crypto/tls
// package of course:
// https://cs.opensource.google/go/go/+/refs/tags/go1.18.4:src/crypto/tls/cipher_suites.go;l=54-72.

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

	if sslCertFlag != "" {
		cnf.Section(`client`).NewKey(`ssl-cert`, sslCertFlag)
	}
	if sslKeyFlag != "" {
		cnf.Section(`client`).NewKey(`ssl-key`, sslKeyFlag)
	}
	if sslCaFlag != "" {
		cnf.Section(`client`).NewKey(`ssl-ca`, sslCaFlag)
	}
}

// Translate cnf to mysql.Config
func cnfToConfig(cnf *ini.File) (*mysql.Config, error) {
	config := mysql.NewConfig()
	if !cnf.HasSection(`client`) {
		return config, nil
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

	// Build network info
	if socket, ok := clientMap[`socket`]; ok {
		config.Net = `unix`
		config.Addr = socket
	} else {
		config.Net = `tcp`

		host, hostok := clientMap[`host`]
		if !hostok {
			host = `127.0.0.1`
		}

		port, portok := clientMap[`port`]
		if !portok {
			port = `3306`
		}
		config.Addr = fmt.Sprintf("%s:%s", host, port)
	}

	// Default connection to 127.0.0.1:3306
	if config.Net == "" {
		config.Addr = `127.0.0.1:3306`

	}

	// SSL Stuff
	var errs *multierror.Error

	// Handle CA
	rootCertPool := x509.NewCertPool()
	if sslca, ok := clientMap[`ssl-ca`]; ok {
		pem, err := os.ReadFile(sslca)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf(`ssl-ca error: %v`, err))
		} else {
			if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
				errs = multierror.Append(errs, errors.New("failed to append PEM"))
			}
		}
	}

	// Handle cert/key
	sslcert, certok := clientMap[`ssl-cert`]
	sslkey, keyok := clientMap[`ssl-key`]
	if (certok && !keyok) || (!certok && keyok) {
		errs = multierror.Append(errs, errors.New("need both ssl-cert and ssl-key set"))
	} else if certok && keyok {
		clientCert := make([]tls.Certificate, 0, 1)
		certs, err := tls.LoadX509KeyPair(sslcert, sslkey)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf(`ssl-cert/key error: %v`, err))
		} else {
			clientCert = append(clientCert, certs)
			mysql.RegisterTLSConfig("custom", &tls.Config{
				RootCAs:      rootCertPool,
				Certificates: clientCert,
			})
			config.TLSConfig = `custom`
		}
	}

	return config, errs.ErrorOrNil()
}
