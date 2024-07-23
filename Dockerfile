# Base image
FROM golang:1.22
 
# Set the Current Working Directory inside the container
WORKDIR /app
 
# Copy go mod and sum files
COPY go.mod go.sum ./
 
# Downloads all dependencies
RUN go mod download
 
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
 
EXPOSE 8080
 
CMD ["go", "run", "main.go"]