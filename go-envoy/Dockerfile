# Start from the official Golang image
FROM golang:1.23

# Set working directory
WORKDIR /app

# Copy source code and compile
COPY . .
RUN go build -o main .

# Expose the app port
EXPOSE 8080

# Run the application
CMD ["./main"]