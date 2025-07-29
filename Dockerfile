FROM golang:1.24.5-alpine

# Set working directory
WORKDIR /app

# Install necessary packages
RUN apk add --no-cache git

# Copy go.mod and download dependencies
COPY go.mod ./
RUN go mod download

# Copy all files
COPY . .

# Build the Go app
RUN go build -o main .

# Run the Go app
CMD ["./main"]
