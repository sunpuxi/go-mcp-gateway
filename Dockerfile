# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o mcp-gateway .

# ---- Stage 2: Run ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/mcp-gateway .
COPY config/config.yaml ./config/config.yaml

EXPOSE 8081

ENTRYPOINT ["./mcp-gateway"]
