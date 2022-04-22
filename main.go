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

var verbose bool // verbose mode
var force bool   // force file/project creation

func usage() {
	fmt.Println("usage: idea [-vy] [project] [file]")
}

func main() {
	args := extractArgs()

	var file string
	var project string

	// no arguments provided, open/create project in the working directory
	if len(args) == 0 {
		project = findOrCreateProjectDir(dod.Getwd())
	}

	// open the given file or project
	if len(args) == 1 {
		path := dod.Abs(args[0])

		if dod.IsDir(path) {
			if isPiped() {
				log.Fatalf("can't pipe into dir %s\n", path)
			}

			project = findOrCreateProjectDir(path)
		} else {
			file = path
			project = findOrCreateProjectDir(dod.Getwd())

			if !dod.PathExists(file) {
				// create none-existing file
				if isPiped() {
					fmt.Printf("creating %s ...\n", file)
				} else if !choice("create " + file + " ?") {
					os.Exit(0)
				}

				dod.CreateFile(file).Close()
			}
		}
	}

	if file != "" {
		startIntelliJ(project, file)
	} else {
		startIntelliJ(project)
	}

	if isPiped() {
		pipeToFile(file)
	}
}

func extractArgs() []string {
	var args []string

	for _, a := range os.Args[1:] {
		if a == "-v" {
			verbose = true
		} else if a == "-f" {
			force = true
		} else {
			args = append(args, a)
		}
	}

	if len(args) > 2 {
		usage()
		log.Fatalf("Too many arguments provided.")
	}

	return args
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

func findOrCreateProjectDir(dir string) string {
	// traverse up the dir tree and locate an IDEA project.
	// if found, open the project instead of the input sub-dir, otherwise IDEA will create a sub-project.
	project := findIdeaDir(dir)

	// project not found up the hierarchy, IDEA will create the .idea dir
	if project == "" {
		project = dir

		if isPiped() {
			fmt.Printf("creating a new IDEA project at %s ...\n", project)
		} else if !choice("create a new IDEA project at " + project + " ?") {
			os.Exit(0)
		}

		dod.CreateDir(project)
	}

	return project
}

func startIntelliJ(paths ...string) {
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
	buf := make([]byte, 0, 4*1024)
	scanner.Buffer(buf, 4*1024)

	for scanner.Scan() {
		fmt.Fprintln(f, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
