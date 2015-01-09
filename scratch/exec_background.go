package main

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
)

func check_binary(exe string) {
	path, err := exec.LookPath(exe)
	if err != nil {
		log.Fatal("installing ", exe, " is in your future")
	}
	log.Print(exe, " is available at ", path)
}

func main() {
	var exe = "mysqladmin"
	cmd := exec.Command(exe, "ext", "-i", "1")

	check_binary(exe)

	// cmd.Stdin = strings.NewReader("some input")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)

	// chan := make(chan string)
	go func() {
		for scanner.Scan() {
			log.Println(scanner.Text()) // Println will add back the final '\n'
			// chan <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Print("error scanning output: ", err)
		}
	}()

	cmd.Wait()

	// fmt.Print(exe, " output: ", out.String())
}
