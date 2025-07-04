# Create Docker network and run LocalStack container
localstack:
	docker run --rm -d \
		-p 4566:4566 \
		-p 4510-4559:4510-4559 \
		--network local-kafka-broker_default \
		--hostname localstack \
		--name localstack \
		localstack/localstack

config:
	@if [ ! -f graphql/schema.json ]; then echo "Error: config/schema.json not found"; exit 1; fi
	@if [ ! -f graphql/connectors.json ]; then echo "Error: config/connectors.json not found"; exit 1; fi
	@if [ ! -f graphql/credentials.json ]; then echo "Error: config/credentials.json not found"; exit 1; fi
	aws --endpoint-url=http://localhost:4566 ssm put-parameter \
		--name "/graphql/dev/schema" \
		--type String \
		--value "$$(cat graphql/schema.json)" \
		--region us-east-1 \
		--overwrite
	aws --endpoint-url=http://localhost:4566 ssm put-parameter \
		--name "/graphql/dev/connectors" \
		--type String \
		--value "$$(cat graphql/connectors.json)" \
		--region us-east-1 \
		--overwrite
	aws --endpoint-url=http://localhost:4566 secretsmanager create-secret \
		--name "/graphql/dev/credentials" \
		--secret-string "$$(cat graphql/credentials.json)" \
		--region us-east-1

# Run locally
local:
	go run .

# Build and run SAM locally with LocalStack integration
run:	
	GOOS=linux GOARCH=amd64 go build -o main .
	sam local start-api \
		--region us-east-1 \
		--docker-network local-kafka-broker_default \
		--warm-containers eager

# Run an image with opentelemetry for local observability tests
otel:
	./setup.sh
	docker compose -f otel-collector-compose.yaml up -d

# Mock values locally with LocalStack SSM
mock:
	@if [ ! -f ../config/mock.json ]; then echo "Error: config/mock.json not found"; exit 1; fi
	aws --endpoint-url=http://localhost:4566 ssm put-parameter \
		--name "/graphql/dev/mock" \
		--type String \
		--value "$$(cat ../config/mock.json)" \
		--region us-east-1 \
		--overwrite

# Clean up
clean:
	-docker stop localstack
	-rm main

.PHONY: localstack config local run otel mock clean