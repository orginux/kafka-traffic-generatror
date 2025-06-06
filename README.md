# Kafka Traffic Generator

This tool generates and sends batches of messages to a Kafka topic using randomly generated data.
Messages are generated in `<key:int><value:json>` format, you can define fields in a config file.

# Usage
Usage of kafka-traffic-generator:
```bash
Environment variables:
  KTG_ACKS string
        Kafka acks setting [all, one, none] (default "none")
  KTG_BATCHDELAY int
        Delay between batches in milliseconds (default "0")
  KTG_BATCHNUM int
        Number of batches (0 - unlimited) (default "0")
  KTG_CA_PATH string
        Path to the CA certificate
  KTG_CERT_PATH string
        Path to the client certificate
  KTG_COMPRESSION string
        Kafka compression setting [none, gzip, snappy, lz4, zstd] (default "none")
  KTG_KAFKA string
        Kafka host address (default "localhost:9092")
  KTG_KEY_PATH string
        Path to the client key
  KTG_LOGLEVEL string
        Logging level [debug, info, warn, error]
  KTG_MSGNUM int
        Number of messages per batch (default "100")
  KTG_TOPIC string
        Kafka topic name
```

## Binary file
#### 1. Build:
```bash
make build
```

#### 2. Create a configuration file in YAML format, e.g., topic.yaml, with the following structure:
```yaml
kafka:
  host: <KAFKA_BROKER_HOST>
  acks: [all, one, none] ("none" by default)
  compression: [gzip, snappy, lz4, zstd] ("none" by default)
loglevel: [debug, information, warning, error]
topic:
  name: <TOPIC_NAME>
  batch_msgs: Positive integer
  batch_count: Positive integer -- 0 — Unlimited number of batches
  batch_delay_ms: Positive integer

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
Example of generating email sending events in a specific time period:
```yaml
---
kafka:
  host: "kafka:9092"
  acks: one
  compression: gzip
topic:
  name: emails
  batch_msgs: 50
  batch_count: 2000
  batch_delay_ms: 500
fields:
  - name: "Date"
    function: daterange
    params:
      format: "yyyy-MM-dd HH:mm:ss"
      startdate: "1993-03-13 15:11:02"
      enddate:  "1993-05-16 15:11:02"
  - name: "Email"
    function: email
    params: {}
  - name: "Message"
    function: sentence
    params: {}
```
Additional examples located within the `./examples` folder, and a comprehensive list of functions is available in [the gofakeit project](https://github.com/brianvoe/gofakeit#functions).

#### 3. Run the program with the path to the configuration file:

```bash
./bin/kafka-traffic-generator --config examples/simple.yaml
```
The program will load the configuration, generate the specified number of messages with random data, and send them to the Kafka topic.

#### 4. After that you can see the messages in your Kafka topic:
```json lines
{"Date":"1993-04-02 17:44:04","Email":"tedvon@carroll.biz","Message":"You with nobody Gabonese my."}
{"Date":"1993-04-20 02:18:18","Email":"ethylmcclure@goldner.info","Message":"By such where deeply so."}
{"Date":"1993-05-08 01:07:46","Email":"betsyoreilly@welch.info","Message":"Firstly of as board she."}
{"Date":"1993-05-08 08:25:17","Email":"theresiapollich@yost.info","Message":"Whom koala scarcely daily how."}
{"Date":"1993-04-28 02:34:36","Email":"colinernser@powlowski.biz","Message":"Other paint yesterday constantly below."}
```

## Docker Image
A Docker image is available for easy deployment of the Kafka Traffic Generator.
To use the Docker image, you can pull it by running the following command:
```bash
docker pull ghcr.io/orginux/kafka-traffic-generator:0.1.128
```

Once you have the image, you can run the Kafka Traffic Generator using Docker Compose.
Here's an example configuration for running the tool:
```yaml
services:
  ktg:
    image: ghcr.io/orginux/kafka-traffic-generator:0.1.128
    container_name: ktg
    networks:
      - kafka-network
    volumes:
      - type: bind
        source: ./configs/
        target: /etc/ktg/
        read_only: true
    environment:
      KTG_CONFIG: /etc/ktg/test_1.yaml
      KTG_KAFKA: "kafka:29092"
      KTG_ACKS: all
      KTG_COMPRESSION: gzip
      KTG_TOPIC: topic1
      KTG_MSGNUM: 10
      KTG_DELAY: 500
      KTG_BATCHNUM: 5
      KTG_LOGLEVEL: debug
```

# ToDo:
- [x] Implement Kafka Authentication: Enhance security by adding Kafka authentication mechanisms;
- [x] Improve Logging: Consider using log/slog to provide better logging capabilities, structured logs, and log levels;
- [x] Add CI/CD: Implement a CI/CD pipeline to automate the build and deployment process;
- [ ] Generate more complex data with dependencies between fields;
- [ ] Add examples for usage shufflestrings;
- [x] Compression;
- [x] Acks setting
- [ ] Implement multithreaded generator;
- [ ] Show final statistics after sending all messages or stop by signal;
- [ ] Implement API: Develop an API for better control and management of the Kafka Traffic Generator;
- [ ] Multi-Topic Support: Extend the generator to support working with multiple Kafka topics simultaneously;
- [ ] Add Tests: Write unit tests to ensure the stability and reliability of the Kafka Traffic Generator;

# Dependencies
This project uses the following Go libraries:
- [brianvoe/gofakeit](https://github.com/brianvoe/gofakeit): A powerful Go library for generating fake data.
- [segmentio/kafka-go](https://github.com/segmentio/kafka-go): A pure Go Kafka client library for interacting with Apache Kafka.
