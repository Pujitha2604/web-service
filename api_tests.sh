#!/bin/bash
 
# Load environment variables from mongo_credentials.txt
export $(grep -v '^#' mongo_credentials.txt | xargs)
 
# Start services with Docker Compose
docker-compose up -d
 
# Wait for the web service to start
echo "Waiting for the web service to start..."
sleep 5  
 
# Run Newman with Docker, mounting the Docker socket and files
docker run --network host \
  --name employee-service \
  -v ./collection/:/etc/postman \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -t postman/newman:latest run -r cli,json --reporter-json-export /etc/postman/newman-report.json /etc/postman/collection.json
 
# Shut down services with Docker Compose
docker-compose down
 