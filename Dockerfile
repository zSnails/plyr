# Build stage
FROM golang:1.21-bookworm AS builder
WORKDIR /app
COPY . .
RUN go build -o plyr .

# Final stage
FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/plyr .
<<<<<<< HEAD

RUN apt-get update && apt-get install -y \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*
COPY creation.sql .
RUN sqlite3 data.sqlite < creation.sql

=======
>>>>>>> c4c4d4db83acb3ac5e1ae917284373a45d10c2cd
EXPOSE 8080
CMD ["./plyr"]
