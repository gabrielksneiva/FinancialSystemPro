FROM golang:1.25 AS builder
WORKDIR /app

# Dependencies para CGO + sqlite3
RUN apt-get update && apt-get install -y gcc g++ libc6-dev sqlite3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown

# Testes com CGO habilitado (cosmo precisa)
RUN CGO_ENABLED=1 go test -failfast ./...

# Build final sem CGO
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "\
    -X financial-system-pro/cmd/server.version=$VERSION \
    -X financial-system-pro/cmd/server.commit=$COMMIT_SHA \
    -X financial-system-pro/cmd/server.buildDate=$BUILD_DATE \
    -s -w" \
    -o main ./cmd/server


FROM alpine:latest
RUN apk --no-cache add ca-certificates postgresql-client tzdata curl
WORKDIR /root/
COPY --from=builder /app/main .
ENV APP_VERSION=$VERSION APP_COMMIT=$COMMIT_SHA APP_BUILD_DATE=$BUILD_DATE
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 CMD curl -f http://localhost:3000/health || exit 1
CMD ["./main"]
