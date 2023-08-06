build:
	go build -o ./bin/kafka-traffic-generator

build-image:
	docker build -t ghcr.io/orginux/kafka-traffic-generatror:latest .

# Tests
test-up: test-down test-start-kafka test-topic-create test-topic-check

test-down:
	docker compose --file tests/docker-compose.yml down
	docker compose --file tests/docker-compose-ktg.yml down

test-start-kafka:
	docker compose --file tests/docker-compose.yml up --remove-orphans -d

test: test-up
	docker compose --file tests/docker-compose-ktg.yml up --exit-code-from ktg

test-local: build-image test


## Kafka
test-topic-create:
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --topic topic1 --create --partitions 6 --replication-factor 1
test-topic-check:
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --describe topic1
test-topic-consumer:
	docker exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --from-beginning --topic topic1
test-topic-lag:
	docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --describe --all-groups

## Generate messages
test-generate: build
	./bin/kafka-traffic-generator --config examples/persons-address.yaml
test-examples:
	bash ./tests/scripts/test-examples.sh
