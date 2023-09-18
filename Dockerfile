# Build stage
FROM golang:bookworm AS builder
WORKDIR /app
COPY . .
RUN go build -o plyr .

# Final stage
FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/plyr .
COPY processed/ ./processed/
COPY data.sqlite .
EXPOSE 8080
CMD ["./plyr"]