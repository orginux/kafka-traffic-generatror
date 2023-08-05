package main

import (
	"flag"
	"log"

	"kafka-traffic-generator/internal/config"
	"kafka-traffic-generator/internal/generator"
)

var (
	configFile string
)

// init initializes the command-line flags.
func init() {
	flag.StringVar(&configFile, "config", "", "config file path")
	flag.Parse()
}

func main() {
	// Load the topic description from a YAML file
	config, err := config.Load(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	// Run the Kafka traffic generator with the provided configuration
	if err = generator.Run(config); err != nil {
		log.Fatalln(err)
	}
}
