package main

import (
	"log"
	"os"

	"github.com/ayanrajpoot10/azsh/internal/cmd"
)

func main() {
	cli := cmd.NewCLI()
	if err := cli.Run(os.Args[1:]); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
