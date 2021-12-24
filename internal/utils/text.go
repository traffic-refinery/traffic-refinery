package utils

import (
	"bufio"
	"os"
)

// GetStringLines reads a file line by line and returns a slice containing one entry per line
func GetStringLines(fname string) []string {
	f, err := os.Open(fname)
	defer f.Close()
	var list []string
	if err != nil {
		return list
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}

	return list
}
