FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the server from its module path (no Go files at repo root)
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown
# Fail fast if tests break before producing binary
RUN go test -failfast ./...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X financial-system-pro/cmd/server.version=$VERSION -X financial-system-pro/cmd/server.commit=$COMMIT_SHA -X financial-system-pro/cmd/server.buildDate=$BUILD_DATE -s -w" -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates postgresql-client tzdata curl
WORKDIR /root/
COPY --from=builder /app/main .
ENV APP_VERSION=$VERSION APP_COMMIT=$COMMIT_SHA APP_BUILD_DATE=$BUILD_DATE
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 CMD curl -f http://localhost:3000/health || exit 1
CMD ["./main"]
