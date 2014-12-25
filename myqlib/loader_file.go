package myqlib

import (
  "os"
  "bufio"
)

// Load mysql status output from a mysqladmin output file
func GetSamplesFile(filename string) (chan MyqSample, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	var ch = make(chan MyqSample)

	// The file scanning goes into the background
	go func() {
		defer file.Close()
		defer close(ch)

		scanMySQLShowLines(scanner, ch)
	}()

	return ch, nil
}
