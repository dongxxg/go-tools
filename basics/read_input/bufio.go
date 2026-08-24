package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter your name:")
	input, err := inputReader.ReadString('\n')
	if err != nil {
		fmt.Println("There were errors, exiting program", err)
		return
	}

	fmt.Printf("Your name is %s", input)

	switch input {
	case "dx\n":
		fmt.Println("Welcome dx!")
	case "shane\n":
		fmt.Println("Welcome shane!")
	default:
		fmt.Println("You are not welcome here, Goodbye!")
	}

}
