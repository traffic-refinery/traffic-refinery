package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetRepoPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Check if already in root directory (e.g. main.go)
	lastSlashIndex := strings.LastIndex(dir, "/")
	cName := dir[lastSlashIndex+1:]
	if cName == "traffic-refinery" {
		return dir
	}

	// Get parent directory
	parent := filepath.Dir(dir)
	lastSlashIndex = strings.LastIndex(parent, "/")
	pName := parent[lastSlashIndex+1:]

	// If not at root, continue getting parent
	for pName != "traffic-refinery" {
		parent = filepath.Dir(parent)
		lastSlashIndex = strings.LastIndex(parent, "/")
		pName = parent[lastSlashIndex+1:]
	}
	return parent
}
