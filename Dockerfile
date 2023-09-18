# Build stage
FROM golang:1.21-bookworm AS builder
WORKDIR /app
COPY . .
RUN go build -o plyr .

# Final stage
FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/plyr .
EXPOSE 8080
CMD ["./plyr"]
