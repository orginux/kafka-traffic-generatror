# Kafka Traffic Generator

This Go program generates and sends batches of messages to a Kafka topic using randomly generated data.

# Usage
1. Build:
``bash
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

3. Run the program with the path to the configuration file:

```bash
./bin/kafka-traffic-generator --config examples/topic.yaml
```
The program will load the configuration, generate the specified number of messages with random data, and send them to the Kafka topic.
