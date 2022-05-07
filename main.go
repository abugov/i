package main

import (
	"bufio"
	"fmt"
	dod "idea/doordie"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var force bool   // force file/project creation
var altproj bool // use the project as an alternate project
var verbose bool // verbose mode

func usage() {
	fmt.Println("usage: idea [-vy] [project] [file]")
}

func main() {
	project, file := parseArgs()

	if project != "" {
		project = dod.Abs(project)
	}

	if file != "" {
		file = dod.Abs(file)
	}

	// try to find a parent project from file
	if project == "" && file != "" {
		project = findIdeaDir(filepath.Dir(file))

		if project == "" {
			project = dod.Getwd()
		}
	}

	if file == "" && isPiped() {
		log.Fatalf(fmt.Sprintf("can't pipe into dir %s\n", project))
	}

	dir := findIdeaDir(project)

	if dir != "" {
		// open the project instead of the input sub-dir (otherwise IDEA will create a subproject).
		project = dir
	} else {
		// project not found, IDEA will create a .idea dir
		if isPiped() {
			fmt.Printf("creating a new IDEA project at %s ...\n", project)
		} else if !choice(fmt.Sprintf("create a new IDEA project at %s ?", project)) {
			os.Exit(0)
		}

		dod.CreateDir(project)
	}

	if file != "" && !dod.PathExists(file) {
		// create none-existing file
		if isPiped() {
			fmt.Printf("creating %s ...\n", file)
		} else if !choice(fmt.Sprintf("create %s ?", file)) {
			os.Exit(0)
		}

		dod.CreateFile(file).Close()
	}

	if file != "" {
		launchIDEA(project, file)
	} else {
		launchIDEA(project)
	}

	if isPiped() {
		pipeToFile(file)
	}
}

func parseArgs() (string, string) {
	var args []string

	for _, a := range os.Args[1:] {
		if a[0] == '-' {
			if strings.Contains(a, "a") {
				altproj = true
			}
			if strings.Contains(a, "v") {
				verbose = true
			}
			if strings.Contains(a, "f") {
				force = true
			}
		} else {
			args = append(args, a)
		}
	}

	if len(args) > 2 {
		usage()
		log.Fatalf("Too many arguments provided.")
	}

	if altproj && len(args) != 2 {
		log.Fatalf("Both project and file are mandatory when using an alternate project.")
	}

	var project string
	var file string

	// no arguments provided, open/create project in the working directory
	if len(args) == 0 {
		project = dod.Getwd()
	}

	// open/create the given file or project
	if len(args) == 1 {
		path := args[0]

		if dod.IsDir(path) {
			project = path
		} else {
			file = path
		}
	}

	if len(args) == 2 {
		project = args[0]
		file = args[1]
	}

	return project, file
}

// return true for "Y", "y" and ""
func choice(message string) bool {
	if force {
		return true
	}

	if isPiped() {
		panic("can't take user choice while piped")
	}

	fmt.Print(message + " [Y/n] ")

	var c string
	fmt.Scanf("%s", &c)
	c = strings.TrimSpace(c)

	return c == "y" || c == "Y" || c == ""
}

// traverse up the dir tree to locate an IDEA project.
func findIdeaDir(path string) string {
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

func launchIDEA(paths ...string) {
	args := []string{"-na", "/Applications/IntelliJ IDEA.app"}

	if len(paths) > 0 {
		args = append(args, "--args")
		args = append(args, paths...)
	}

	if verbose {
		fmt.Printf("/usr/bin/open %v\n", strings.Join(args, " "))
	}

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
	return fi.Mode()&os.ModeCharDevice == 0
}

func pipeToFile(file string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// TODO: check cmd.StdinPipe()
	scanner := bufio.NewScanner(os.Stdin)
	//buf := make([]byte, 0, 4*1024)
	//scanner.Buffer(buf, 4*1024)

	for scanner.Scan() {
		fmt.Fprintln(f, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
