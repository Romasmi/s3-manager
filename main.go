package main

import (
	"log"
	"s3manager/config"
)

func main() {
	cnf, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println(cnf.ApiURL)
}
