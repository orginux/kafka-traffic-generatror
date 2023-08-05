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

	if err = generator.Run(config); err != nil {
		log.Fatalln(err)
	}
}
