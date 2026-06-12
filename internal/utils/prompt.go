package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func PromptInput(prompt string) (string, error) {
	fmt.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func PromptSelect(prompt string, options []string) (int, error) {
	fmt.Println(prompt)
	for i, opt := range options {
		fmt.Printf("  %2d) %s\n", i+1, opt)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter number: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		input = strings.TrimSpace(input)
		index, err := strconv.Atoi(input)
		if err == nil && index >= 1 && index <= len(options) {
			return index - 1, nil
		}
		fmt.Printf("Invalid selection. Enter a number between 1 and %d.\n", len(options))
	}
}
