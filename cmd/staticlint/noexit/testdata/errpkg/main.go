package main

import "os"

func main() {
	os.Exit(0) // want "direct call to os.Exit from main function is forbidden"
}
