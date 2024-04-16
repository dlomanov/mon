package main

import "os"

func main() {
	os.Exit(0) // want "main function should not exit via os.Exit"
}
