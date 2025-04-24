package main

import (
	"github.com/microcodebase/microconfig"
	"log"
)

var config map[string]string

func main() {
	var err error
	config, err = microconfig.ParseFile("config.txt")
	if err != nil {
		log.Println("Failed to parse config", err)
		return
	}

	Extractor()
}
