package main

import (
    "bufio"
    "os/exec"
    "fmt"
    "log"
    "os"
    "time"
    "errors"
    "strings"
    "os/user"
    "path/filepath"
)

func main() {
    args := os.Args[1:]

    usr, _ := user.Current()
    // use ~/documents/new project for temp files and files without a project
    docsNewDir := filepath.Join(usr.HomeDir, "Documents/new")

    var path string
    var project string

    // no arguments provided, create a temp file and open the "new" project in a separate instance
    if len(args) == 0 {
        // TODO: try -t (Opens with default text editor) so .txt can be removed
        path = filepath.Join(docsNewDir, time.Now().Format("2006-01-02_15:04:05") + ".txt")
        createFileOrDie(path).Close()
    }

    // open the given file or project
    if len(args) == 1 {
        var err error
        path, err = filepath.Abs(args[0])

        if err != nil {
            log.Fatal(err)
        }

        // create none-existing file
        if !pathExistsOrDie(path) {
            if isPiped() {
                fmt.Printf("creating %s ...\n", path)
            } else if !choice("create " + path + " ?") {
                os.Exit(0)
            }

            createFileOrDie(path).Close()
        }

        // traverse up the dir tree and locate an idea project
        project = findIdea(path)

        if project != "" {
            if isDirOrDie(path) {
                // open the project instead of the input sub-dir, otherwise idea will create a sub-project
                path = project
                project = ""
            }
        } else if isDirOrDie(path) {
            if isPiped() {
                log.Fatalf("can't pipe into dir %s\n", path)
            }

            // a single dir was provided but no project was found up the hierarchy, intellij will create .idea dir
            if !choice("create new idea project at " + path + " ?") {
                os.Exit(0)
            }
        } else {
            // a single file without a parent project was provided, use the ~/documents/new project for files without project
            project = docsNewDir
        }
    }

    if project != "" {
        startIntelliJ(project, path)
    } else {
        startIntelliJ(path)
    }

    if isPiped() {
        if isDirOrDie(path) {
            log.Fatalf("can't pipe into dir %s\n", path)
        }

        pipeToFile(path)
    }
}

func createFileOrDie(file string) *os.File {
    f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

    if err != nil {
        log.Fatal(err)
    }

    return f
}

func pathExistsOrDie(path string) bool {
    if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
      return false
    } else if err != nil {
        log.Fatal(err)
    }
    return true
}

func isDirOrDie(dir string) bool {
    if f, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
      return false
    } else if err != nil {
        log.Fatal(err)
    } else {
        return f.IsDir()
    }
    return false
}

// return true for "Y", "y" and ""
func choice(message string) bool {
    fmt.Print(message + " [Y/n] ")

    var c string
    fmt.Scanf("%s", &c)
    c = strings.ToLower(strings.TrimSpace(c))

    return c == "y" || c == ""
}

func findIdea(path string) string {
    if !pathExistsOrDie(path) {
        return ""
    }

    if !isDirOrDie(path) {
        path = filepath.Dir(path)
    }

    for true {
        idea := filepath.Join(path, ".idea")

        if isDirOrDie(idea) {
            return path
        }

        if path == "/" {
            return ""
        }

        path = filepath.Dir(path)
    }

    return ""
}

func startIntelliJ(paths ...string) {
    args := []string{"-na", "/Applications/IntelliJ IDEA.app"}

    if len(paths) > 0 {
        args = append(args, "--args")
        args = append(args, paths...)
    }

    //fmt.Printf("/usr/bin/open %v\n", strings.Join(args, " "))

    cmd := exec.Command("/usr/bin/open", args...)
    err := cmd.Run()

    b, _ := cmd.CombinedOutput()

    if b != nil {
        fmt.Println(string(b))
    }

    if err != nil {
        log.Fatal(err)
    }
}

func isPiped() bool {
    fi, _ := os.Stdin.Stat()
    return (fi.Mode() & os.ModeCharDevice) == 0
}

func pipeToFile(file string) {
    f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // TODO: check cmd.StdinPipe()
    scanner := bufio.NewScanner(os.Stdin)
    buf := make([]byte, 0, 4*1024)
    scanner.Buffer(buf, 4*1024)

    for scanner.Scan() {
        fmt.Fprintln(f, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
}