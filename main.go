package main

import (
	"log"

	"kafka-traffic-generator/internal/config"
	"kafka-traffic-generator/internal/generator"
)

func main() {
	// Load the topic description from a YAML file
	config, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	// Run the Kafka traffic generator with the provided configuration
	if err = generator.Run(*config); err != nil {
		log.Fatalln(err)
	}
}
