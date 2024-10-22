package main

import (
	"fmt"
	"os"

	"github.com/sirrobot01/lamba/cmd"
)

func main() {
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
