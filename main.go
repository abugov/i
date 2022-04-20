package main

import (
    "bufio"
    "os/exec"
    "fmt"
    "log"
    "os"
    "time"
    "strings"
    "os/user"
    "path/filepath"
    dod "i/doordie"
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
        project = docsNewDir

        if !dod.PathExists(docsNewDir) {
            fmt.Printf("creating idea project %s ...\n", docsNewDir)
            dod.CreateDir(docsNewDir)
        }

        dod.CreateFile(path).Close()
    }

    // open the given file or project
    if len(args) == 1 {
        var err error
        path, err = filepath.Abs(args[0])

        if err != nil {
            log.Fatal(err)
        }

        // create none-existing file
        if !dod.PathExists(path) {
            if isPiped() {
                fmt.Printf("creating %s ...\n", path)
            } else if !choice("create " + path + " ?") {
                os.Exit(0)
            }

            dod.CreateFile(path).Close()
        }

        // traverse up the dir tree and locate an idea project
        project = findIdea(path)

        if project != "" {
            if dod.IsDir(path) {
                // open the project instead of the input sub-dir, otherwise idea will create a sub-project
                path = project
                project = ""
            }
        } else if dod.IsDir(path) {
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

            if !dod.PathExists(docsNewDir) {
                fmt.Printf("creating idea project %s ...\n", docsNewDir)
                dod.CreateDir(docsNewDir)
            }
        }
    }

    if project != "" {
        startIntelliJ(project, path)
    } else {
        startIntelliJ(path)
    }

    if isPiped() {
        if dod.IsDir(path) {
            log.Fatalf("can't pipe into dir %s\n", path)
        }

        pipeToFile(path)
    }
}

// return true for "Y", "y" and ""
func choice(message string) bool {
    if isPiped() {
        panic("can't take user choice while piped")
    }

    fmt.Print(message + " [Y/n] ")

    var c string
    fmt.Scanf("%s", &c)
    c = strings.TrimSpace(c)

    return c == "y" || c == "Y" || c == ""
}

func findIdea(path string) string {
    if !dod.PathExists(path) {
        return ""
    }

    if !dod.IsDir(path) {
        path = filepath.Dir(path)
    }

    for true {
        idea := filepath.Join(path, ".idea")

        if dod.IsDir(idea) {
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

    fmt.Printf("/usr/bin/open %v\n", strings.Join(args, " "))

    cmd := exec.Command("/usr/bin/open", args...)
    err := cmd.Run()

    if b, _ := cmd.CombinedOutput(); b != nil {
        fmt.Println(string(b))
    }

    if err != nil {
        log.Fatal(err)
    }
}

func isPiped() bool {
    fi, _ := os.Stdin.Stat()
    return fi.Mode() & os.ModeCharDevice == 0
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