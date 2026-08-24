package main

import "fmt"

var (
	firstName, lastName, s string
	i                      int
	f                      float32
	input                  = "56.2 / 5211 / Go"
	format                 = "%f / %d / %s"
)

func main() {
	fmt.Println("Please enter your full name:")
	fmt.Scanln(&firstName, &lastName)
	fmt.Printf("Hi, %s %s\n", lastName, firstName)
	fmt.Sscanf(input, format, &f, &i, &s)
	fmt.Println("From the string we read:", f, i, s)
}
