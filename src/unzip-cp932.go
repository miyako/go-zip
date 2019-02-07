package main

import (
	"./zi18np"
	"fmt"
	"os"
	"path/filepath"
)

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	if 2 > len(os.Args) {
		fmt.Printf("usage:\n\t%s [zip file]\n", os.Args[0])
		os.Exit(1)
	}
	file := os.Args[1]

	if !Exists(file) {
		fmt.Printf("%s not found.\n", file)
		os.Exit(1)
	}
	err := zi18np.Unzip(file, filepath.Dir(file))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
