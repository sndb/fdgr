package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func status(dirty, clean *int) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" {
			dir := filepath.Dir(path)
			cmd := exec.Command("git", "status", "-s")
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if len(out) > 0 {
				fmt.Printf("%q:\n%s", dir, colorize(colorRed, string(out)))
				*dirty++
			} else {
				fmt.Printf("%q: %v\n", dir, colorize(colorGreen, "clean"))
				*clean++
			}
			if err != nil {
				return err
			}
			return fs.SkipDir
		}
		return nil
	}
}

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dirs = append(dirs, dir)
	}
	for _, d := range dirs {
		var dirty, clean int
		if err := filepath.WalkDir(d, status(&dirty, &clean)); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("walked %q: %v dirty, %v clean\n", d, dirty, clean)
	}
}
