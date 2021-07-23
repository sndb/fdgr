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

func checkFile(info *walkInfo) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i := sort.SearchStrings(info.ignoreDirs, d.Name())
		if i < len(info.ignoreDirs) && info.ignoreDirs[i] == d.Name() {
			return fs.SkipDir
		}
		if d.Name() == ".git" {
			dir := filepath.Dir(path)
			statusCmd := exec.Command("git", "status", "-s")
			statusCmd.Dir = dir
			out, err := statusCmd.CombinedOutput()
			if len(out) > 0 {
				fmt.Printf("%q:\n%s", dir, colorize(colorRed, string(out)))
				info.dirty++
			} else {
				fmt.Printf("%q: %v\n", dir, colorize(colorGreen, "clean"))
				info.clean++
			}
			if err != nil {
				return err
			}
			return fs.SkipDir
		}
		return nil
	}
}

type walkInfo struct {
	dirty int
	clean int
	ignoreDirs
}

type ignoreDirs []string

func (l *ignoreDirs) String() string {
	return strings.Join(*l, ", ")
}

func (l *ignoreDirs) Set(s string) error {
	*l = strings.Split(s, ",")
	sort.StringSlice(*l).Sort()
	return nil
}

func main() {
	var ignore ignoreDirs
	flag.Var(&ignore, "ignore", "comma-separated list of directories to ignore")
	flag.Parse()

	dirs := flag.Args()
	if len(dirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dirs = append(dirs, cwd)
	}

	for _, dir := range dirs {
		info := &walkInfo{ignoreDirs: ignore}
		if err := filepath.WalkDir(dir, checkFile(info)); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("walked %q: %v dirty, %v clean\n", dir, info.dirty, info.clean)
	}
}
