# Kafka Traffic Generator

This Go program generates and sends batches of messages to a Kafka topic using randomly generated data.
Messages are generated in <key><json> format, you can define fields in a config file.

# Usage
1. Build:
```bash
make build
```

2. Create a configuration file in YAML format, e.g., topic.yaml, with the following structure:
```yaml
kafka:
  host: <KAFKA_BROKER_HOST>

topic:
  name: <TOPIC_NAME>
  batch_msgs: <NUMBER_OF_MESSAGES_PER_BATCH>
  batch_count: <NUMBER_OF_BATCHES>
  batch_delay_ms: <DELAY_BETWEEN_BATCHES_IN_MILLISECONDS>

fields:
  - name: <FIELD_NAME_1>
    function: <FIELD_GENERATION_FUNCTION_1>
    params:
      <PARAMETER_1>: <VALUE_1>
      <PARAMETER_2>: <VALUE_2>
      ...

  - name: <FIELD_NAME_2>
    function: <FIELD_GENERATION_FUNCTION_2>
    params:
      <PARAMETER_1>: <VALUE_1>
      <PARAMETER_2>: <VALUE_2>
      ...
```
All functions are listed in [the gofakeit project](https://github.com/brianvoe/gofakeit#functions).

3. Run the program with the path to the configuration file:

```bash
./bin/kafka-traffic-generator --config examples/simple.yaml
```
The program will load the configuration, generate the specified number of messages with random data, and send them to the Kafka topic.

# Dependencies
This project uses the following Go libraries:
- [brianvoe/gofakeit](https://github.com/brianvoe/gofakeit): A powerful Go library for generating fake data.
- [segmentio/kafka-go](https://github.com/segmentio/kafka-go): A pure Go Kafka client library for interacting with Apache Kafka.
