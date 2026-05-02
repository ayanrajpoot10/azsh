package main

import (
	"fmt"
	"os"

	"github.com/ayanrajpoot10/azsh/internal/cli"
)

func main() {
	c := cli.New()
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Println("Error: %v", err)
		os.Exit(1)
	}
}
