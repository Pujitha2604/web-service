#!/bin/bash
export $(grep -v '^#' mongo_credentials.txt | xargs)
# Ensure the test containers are stopped before starting
docker-compose down

# Start the test containers
docker-compose up -d

# Run tests with coverage
go test ./... -coverprofile=coverage.out


# Display coverage
go tool cover -html=coverage.out -o coverage.html

# Open coverage report
xdg-open coverage.html

# Stop the test containers
docker-compose down
