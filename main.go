package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorWhite  = "\033[97m"
)

func colorize(color, s string) string {
	if runtime.GOOS == "windows" {
		return s
	}
	return color + s + colorReset
}

type walkInfo struct {
	dirty int
	clean int
	ignoreDirs
}

func (wi *walkInfo) walkDirFunc() fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i := sort.SearchStrings(wi.ignoreDirs, d.Name())
		if i < len(wi.ignoreDirs) && wi.ignoreDirs[i] == d.Name() {
			return fs.SkipDir
		}
		if d.Name() == ".git" {
			dir := filepath.Dir(path)
			cmd := exec.Command("git", "status", "-s")
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if len(out) > 0 {
				fmt.Printf("%q:\n%s", dir, colorize(colorRed, string(out)))
				wi.dirty++
			} else {
				fmt.Printf("%q: %v\n", dir, colorize(colorGreen, "clean"))
				wi.clean++
			}
			if err != nil {
				return err
			}
			return fs.SkipDir
		}
		return nil
	}
}

type ignoreDirs []string

func (d *ignoreDirs) String() string {
	return strings.Join(*d, ", ")
}

func (d *ignoreDirs) Set(s string) error {
	*d = strings.Split(s, ",")
	sort.StringSlice(*d).Sort()
	return nil
}

func main() {
	var ids ignoreDirs
	flag.Var(&ids, "ignore", "comma-separated list of directories to ignore")
	flag.Parse()

	dirs := flag.Args()
	if len(dirs) == 0 {
		d, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dirs = append(dirs, d)
	}

	for _, d := range dirs {
		wi := &walkInfo{ignoreDirs: ids}
		if err := filepath.WalkDir(d, wi.walkDirFunc()); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("walked %q: %v dirty, %v clean\n", d, wi.dirty, wi.clean)
	}
}
