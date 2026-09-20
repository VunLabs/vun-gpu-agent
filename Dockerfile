FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o vun-agent ./cmd/agent

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/vun-agent /usr/local/bin/vun-agent

ENTRYPOINT ["vun-agent"]