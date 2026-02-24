package main

import (
	"log"

	"lewist543.com/mass-project-updater/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
