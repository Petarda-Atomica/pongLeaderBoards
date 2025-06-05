package main

import (
	"bufio"
	"os/exec"
	"strings"
)

func bore(port string) (string, error) {
	// Create command
	cmd := exec.Command("bore", "local", port, "--to", "bore.pub")

	// Pipe stdout to a variable
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	// Start command
	err = cmd.Start()
	if err != nil {
		return "", err
	}

	// Create a scanner to read stdout line-by-line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "listening at") {
			elems := strings.Split(text, " ")
			return elems[len(elems)-1], nil
		}
	}

	return "", nil
}
