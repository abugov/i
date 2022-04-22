package doordir

// "Do or Die ..." helpers

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CreateFile(file string) *os.File {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatal(err)
	}

	return f
}

func CreateDir(dir string) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		log.Fatal(err)
	}
	return true
}

func IsDir(dir string) bool {
	if f, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		log.Fatal(err)
	} else {
		return f.IsDir()
	}
	return false
}

func Getwd() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return path
}

func Abs(path string) string {
	if strings.Count(path, "~") > 1 {
		log.Fatal("Invalid path: " + path)
	} else if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		path = filepath.Join(home, path[2:])
	}

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	return path
}
