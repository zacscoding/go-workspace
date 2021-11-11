package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/manifoldco/promptui"
)

func main() {
	command := "prompt"
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}

	switch command {
	case "prompt":
		runPrompt()
	case "select":
		runSelect()
	default:
		log.Fatal("unknown command: " + os.Args[1])
	}
}

func runPrompt() {
	validate := func(input string) error {
		_, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return errors.New("invalid number")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Number",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	fmt.Printf("You choose %q\n", result)
}

func runSelect() {
	prompt := promptui.Select{
		Label: "Select Day",
		Items: []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday",
			"Saturday", "Sunday"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", result)
}