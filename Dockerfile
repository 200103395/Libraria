# Use an official Golang runtime as the base image
FROM golang:1.22 AS builder

# Set the working directory in the container
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code to the working directory
COPY . ./

RUN apt-get update
RUN apt-get -y install postgresql-client

RUN chmod +x ./wait-for-postgres.sh

# Build the Go application
RUN go build -o app .

# Expose port 8000 to the outside world
EXPOSE 8000

# Command to run the executable
CMD ["/app/app"]
