package main

import (
	"fmt"
	"generator/constant"
	"generator/libgen"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Use: fsagen <path> <limit in numbers> [sub folder depth]")
		return
	}

	limit, err := strconv.ParseInt(os.Args[2], 10, 64)
	handle(err)

	path := os.Args[1]
	absPath, err := getRootPath(path)
	handle(err)

	var depth = constant.DefaultDepth
	if len(os.Args) > 3 {
		depth, err = strconv.ParseInt(os.Args[3], 10, 64)
		handle(err)
	}

	filecount := 5 * limit * limit * depth
	fmt.Printf("Generating %d files.....", filecount)

	err = libgen.GenerateFiles(absPath, limit, depth)
	handle(err)

	fmt.Println("Done!")
}

func getRootPath(path string) (string, error) {
	finfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, fs.ModePerm)
		if err != nil {
			return "", err
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if finfo != nil {
		if !finfo.IsDir() {
			return filepath.Dir(absPath), nil
		}
	}

	return absPath, nil
}

func handle(err error) {
	if err != nil {
		fmt.Printf("\n\n%v\n\n", err)
		os.Exit(1)
	}
}
