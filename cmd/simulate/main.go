package main

import (
	"log"
	"os"
)

func main() {
	log.Println("Simulation command is disabled in this build.")
	log.Println("To re-enable, move shared logic into an importable package and update this command accordingly.")
	os.Exit(1)
}
