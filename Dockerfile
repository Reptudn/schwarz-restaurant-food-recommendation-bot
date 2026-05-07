FROM golang:1.23-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o food-recommender .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    chromium \
    ca-certificates \
    fonts-liberation \
    --no-install-recommends && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/food-recommender .

ENV CHROMIUM_PATH=/usr/bin/chromium

CMD ["./food-recommender"]