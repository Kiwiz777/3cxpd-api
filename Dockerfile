#build stage
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Stage 2: Production
FROM debian:bullseye-slim
ENV GO_ENV=production \
    PORT=3000
# Set the Current Working Directory inside the container
WORKDIR /app
# Copy the prebuilt binary from the builder stage
COPY --from=builder /app/main .

LABEL Name=spotappgo Version=0.0.1
# Expose the port the app runs on
EXPOSE 3000
# Command to run the executable
CMD ["./main"]