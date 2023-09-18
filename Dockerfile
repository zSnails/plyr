# Build stage
FROM golang:1.21-bookworm AS builder
WORKDIR /app
COPY . .
RUN go build -o plyr .

# Final stage
FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/plyr .
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*
COPY creation.sql .
RUN sqlite3 data.sqlite < creation.sql
RUN mkdir /app/songs
EXPOSE 8080
CMD ["./plyr"]