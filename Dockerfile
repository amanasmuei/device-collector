# Start from the official golang image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Set the timezone as an environment variable
ENV TZ=Asia/Kuala_Lumpur

# Install the tzdata package (specific to Debian/Ubuntu-based images)
RUN apt-get update && apt-get install -y tzdata

# Set the timezone
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]