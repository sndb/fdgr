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
	ignoredDirs
}

func newWalkInfo(i ignoredDirs) *walkInfo {
	return &walkInfo{
		ignoredDirs: i,
	}
}

func (i *walkInfo) walkDirFunc() fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if i.ignoredDirs.check(d.Name()) {
			return fs.SkipDir
		}

		if d.Name() == ".git" {
			dir := filepath.Dir(path)
			cmd := exec.Command("git", "status", "-s")
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if len(out) > 0 {
				fmt.Printf("%q:\n%s", dir, colorize(colorRed, string(out)))
				i.dirty++
			} else {
				fmt.Printf("%q: %v\n", dir, colorize(colorGreen, "clean"))
				i.clean++
			}
			if err != nil {
				return err
			}
			return fs.SkipDir
		}
		return nil
	}
}

type ignoredDirs []string

func (i *ignoredDirs) String() string {
	return strings.Join(*i, ",")
}

func (i *ignoredDirs) Set(s string) error {
	*i = strings.Split(s, ",")
	sort.StringSlice(*i).Sort()
	return nil
}

func (i *ignoredDirs) check(dir string) bool {
	j := sort.SearchStrings(*i, dir)
	if j < len(*i) && (*i)[j] == dir {
		return true
	}
	return false
}

func main() {
	var ignored ignoredDirs
	flag.Var(&ignored, "ignore", "comma-separated list of directories to ignore")
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
		wi := newWalkInfo(ignored)
		if err := filepath.WalkDir(d, wi.walkDirFunc()); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("walked %q: %v dirty, %v clean\n", d, wi.dirty, wi.clean)
	}
}
