package doordir

// "Do or Die ..." helpers

import (
    "log"
    "os"
    "errors"
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

