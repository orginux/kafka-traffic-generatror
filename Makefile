up:
	docker compose up -d

down:
	docker compose down

# Kafka
topic-create:
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --topic topic1 --create --partitions 6 --replication-factor 1
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --topic topic2 --create --partitions 6 --replication-factor 1
topic-check:
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --describe topic1
	docker exec kafka kafka-topics --bootstrap-server kafka:9092 --describe topic2

topic-consumer:
	docker exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic topic1
topic-lag:
	docker exec kafka kafka-run-class kafka.admin.ConsumerGroupCommand --group group_2 --bootstrap-server kafka:9092 --describe

# Prepare lab
create: up topic-create topic-check
