package main

import (
	"log"
	"os"
	"s3manager/cmd"
	"s3manager/config"
)

func main() {
	cnf, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration " + err.Error())
	}
	if err := cmd.Execute(cnf); err != nil {
		log.Printf("Failed to execute command " + err.Error())
		os.Exit(1)
	}
}
