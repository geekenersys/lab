.PHONY: build run test docker-build docker-run clean dev db-up db-down

# Project settings
BINARY_NAME=myapp
DOCKER_IMAGE=myapp-image

# Build the project
build:
	@echo "Building..."
	go build -o ${BINARY_NAME}

# Run the project
run: build
	@echo "Running..."
	./${BINARY_NAME}

# Run the project with air for live reloading
dev:
	@echo "Running with air for live reloading..."
	air

# Test the project
test:
	@echo "Testing..."
	go test ./...

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t ${DOCKER_IMAGE} .

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 ${DOCKER_IMAGE}

# Clean up
clean:
	@echo "Cleaning..."
	go clean
	rm ${BINARY_NAME}

# Database operations (assuming Docker Compose is used for DB)
db-up:
	docker-compose -f ./infra/docker-compose.yml up -d

db-down:
	docker-compose -f ./infra/docker-compose.yml down
