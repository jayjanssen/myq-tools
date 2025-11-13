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

	// Add the --defaults-file if it was given
	if defaultsFile != "" {
		files = append(files, defaultsFile)
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

// Append each of the given files to the cnf
func appendFiles(cnf *ini.File, files []string) error {
	var errs *multierror.Error

	for _, file := range files {
		err := cnf.Append(file)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs.ErrorOrNil()
}

// Command line flags
var defaultsFile string
var userFlag string
var passwordFlag string
var hostFlag string
var portFlag string
var socketFlag string
var sslCertFlag string
var sslKeyFlag string
var sslCaFlag string
var sslMode string
var enableCleartextPlugin bool

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
	if sslMode != "" {
		cnf.Section(`client`).NewKey(`ssl-mode`, sslMode)
	}

	if enableCleartextPlugin {
		cnf.Section(`client`).NewBooleanKey(`enable-cleartext-plugin`)
	}

}

// getConfigValue looks up a key in the clientMap, checking for both the standard
// key name and the loose- prefixed version. The standard key takes precedence.
// This implements MySQL's loose- prefix behavior where options prefixed with loose-
// are processed normally but won't cause errors if unrecognized.
func getConfigValue(clientMap map[string]string, key string) (string, bool) {
	if val, ok := clientMap[key]; ok {
		return val, true
	}
	if val, ok := clientMap[`loose-`+key]; ok {
		return val, true
	}
	return "", false
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
	if cnfval, ok := getConfigValue(clientMap, `user`); ok {
		config.User = cnfval
	}
	if cnfval, ok := getConfigValue(clientMap, `password`); ok {
		config.Passwd = cnfval
	}

	// Build network info
	if socket, ok := getConfigValue(clientMap, `socket`); ok {
		config.Net = `unix`
		config.Addr = socket
	} else {
		config.Net = `tcp`

		host, hostok := getConfigValue(clientMap, `host`)
		if !hostok {
			host = `127.0.0.1`
		}

		port, portok := getConfigValue(clientMap, `port`)
		if !portok {
			port = `3306`
		}
		config.Addr = fmt.Sprintf("%s:%s", host, port)
	}

	// Default connection to 127.0.0.1:3306
	if config.Net == "" {
		config.Addr = `127.0.0.1:3306`
	}

	if _, ok := getConfigValue(clientMap, `enable-cleartext-plugin`); ok {
		config.AllowCleartextPasswords = true
	}

	// SSL Stuff
	var errs *multierror.Error
	TLSConfig := &tls.Config{}
	useTLS := false

	// Handle SSL mode
	if sslmode, ok := getConfigValue(clientMap, `ssl-mode`); ok {
		switch sslmode {
		case `VERIFY_CA`: // CA only
			TLSConfig.InsecureSkipVerify = true
			useTLS = true
		}
	}

	// Handle CA
	rootCertPool := x509.NewCertPool()
	if sslca, ok := getConfigValue(clientMap, `ssl-ca`); ok {
		pem, err := os.ReadFile(sslca)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf(`ssl-ca error: %v`, err))
		} else {
			if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
				errs = multierror.Append(errs, errors.New("failed to append PEM"))
			} else {
				TLSConfig.RootCAs = rootCertPool
				useTLS = true
			}
		}
	}

	// Handle cert/key
	sslcert, certok := getConfigValue(clientMap, `ssl-cert`)
	sslkey, keyok := getConfigValue(clientMap, `ssl-key`)
	if (certok && !keyok) || (!certok && keyok) {
		errs = multierror.Append(errs, errors.New("need both ssl-cert and ssl-key set"))
	} else if certok && keyok {
		clientCert := make([]tls.Certificate, 0, 1)
		certs, err := tls.LoadX509KeyPair(sslcert, sslkey)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf(`ssl-cert/key error: %v`, err))
		} else {
			clientCert = append(clientCert, certs)
			TLSConfig.Certificates = clientCert
			useTLS = true
		}
	}

	if useTLS {
		mysql.RegisterTLSConfig("custom", TLSConfig)
		config.TLSConfig = `custom`
	}

	return config, errs.ErrorOrNil()
}
