FROM golang:1.25-alpine AS builder
WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ipupdater .

FROM alpine:latest AS run

RUN apk --no-cache add ca-certificates

RUN addgroup -S appuser && adduser -S -G appuser -H -s /sbin/nologin appuser

WORKDIR /app

COPY --from=builder --chown appuser:appuser /app/ipupdater .

USER appuser

ENTRYPOINT ["/app/ipupdater"]

#CMD ["./ipupdater"]
