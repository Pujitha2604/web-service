# Load environment variables from mongo_credentials.txt
Get-Content mongo_credentials.txt | ForEach-Object {
    $parts = $_ -split '='
    if ($parts.Length -eq 2) {
        [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1], [System.EnvironmentVariableTarget]::Process)
    }
}

# Start services with Docker Compose
docker-compose up -d

# Wait for the web service to start
Write-Output "Waiting for the web service to start..."
Start-Sleep -Seconds 15

# Run Newman with Docker, mounting the Docker socket and files
docker run --network host `
  --name employee-service `
  -v C:/Users/Rekanto/Desktop/employee-service/collection/:/etc/postman `
  -v /var/run/docker.sock:/var/run/docker.sock `
  -t postman/newman:latest run -r cli,json --reporter-json-export /etc/postman/newman-report.json /etc/postman/collection.json

# Shut down services with Docker Compose
docker-compose down
