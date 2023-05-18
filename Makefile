up:
	docker compose up -d

down:
	docker compose down

# Kafka
topic-create:
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --topic topic1 --create --partitions 6 --replication-factor 1
topic-check:
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --describe topic1
topic-consumer:
	docker exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --from-beginning --topic topic1
topic-lag:
	docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --describe --all-groups

# Prepare lab
create: up topic-create topic-check
