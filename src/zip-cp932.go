package main

import (
	"fmt"
	"os"
	"zi18np"
)

func main() {
	if 3 > len(os.Args) {
		fmt.Printf("usage:\n\t%s [zip file] [archive directory]\n", os.Args[0])
		os.Exit(1)
	}

	output := os.Args[1]
	source := os.Args[2]

	if err := zi18np.Zip(source, output); err != nil {
		panic(err)
	}
}
