package main

import (
	"log"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"gopkg.in/yaml.v2"
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

// loadConfig loads the configuration from a YAML file.
func loadConfig(filename string) (Config, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("Error reading YAML file: %v\n", err)
	}

	var config Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return Config{}, fmt.Errorf("Error parsing YAML file: %v\n", err)
	}

	return config, nil
}

// generateBatch generates a batch of Kafka messages with random key-value pairs.
func generateFields(fieldConfigs []Field) []gofakeit.Field {
	var fields []gofakeit.Field
	for _, fieldConfig := range fieldConfigs {
		params := gofakeit.NewMapParams()
		for key, value := range fieldConfig.Params {
			params.Add(key, value)
		}

		field := gofakeit.Field{
			Name:     fieldConfig.Name,
			Function: fieldConfig.Function,
			Params:   *params,
		}

		fields = append(fields, field)
	}

	return fields
}

// sendBatch sends a batch of Kafka messages to the specified topic.
func generateBatch(numMsgs int, fields []gofakeit.Field) ([]kafka.Message, error) {
	var batch []kafka.Message
	for i := 0; i < numMsgs; i++ {
		// TODO make it optional
		key := strconv.Itoa(rand.Intn(100))

		// Generate the random fake data
		jo := gofakeit.JSONOptions{
			Type:   "object",
			Fields: fields,
			Indent: false,
		}
		value, err := gofakeit.JSON(&jo)
		if err != nil {
			return nil, fmt.Errorf("Error of generate random data: %v\n", err)
		}

		// Prepare a Kafka message with the random data
		msg := kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		}
		batch = append(batch, msg)
	}
	return batch, nil
}

func sendBatch(host, topic string, batch []kafka.Message) error {
	mechanism, err := scram.Mechanism(scram.SHA512, "ktg", "1212")
	if err != nil {
		return err
	}

	sharedTransport := &kafka.Transport{
		SASL: mechanism,
	}

	conn := kafka.Writer{
		Addr:      kafka.TCP(host),
		Topic:     topic,
		Transport: sharedTransport,
	}

	err = conn.WriteMessages(context.Background(), batch...)
	if err != nil {
		return fmt.Errorf("Failed to write messages: %v\n", err)
	}
	log.Printf("Sent batch of %d messages\n", len(batch))

	if err := conn.Close(); err != nil {
		return fmt.Errorf("Failed to close the Kafka connection: %v\n", err)
	}

	return nil
}
