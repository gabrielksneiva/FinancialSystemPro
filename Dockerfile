###############################################
# STAGE 1 — Builder com CGO habilitado
###############################################
FROM golang:1.25 AS builder

WORKDIR /app

# Dependências necessárias para CGO + SQLite
RUN apt-get update && apt-get install -y gcc g++ libc6-dev sqlite3 && \
    apt-get clean

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build args (visíveis só neste estágio)
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown

# Executa TODOS os testes com CGO habilitado
RUN CGO_ENABLED=1 go test -failfast ./...

# Binário final — estático
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "\
    -X financial-system-pro/cmd/server.version=$VERSION \
    -X financial-system-pro/cmd/server.commit=$COMMIT_SHA \
    -X financial-system-pro/cmd/server.buildDate=$BUILD_DATE \
    -s -w" \
    -o main ./cmd/server


###############################################
# STAGE 2 — Runtime super leve
###############################################
FROM alpine:latest

# Para runtime somente
RUN apk --no-cache add ca-certificates tzdata curl

# Build args repetidos (necessário para ENV ler)
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown

ENV APP_VERSION=${VERSION}
ENV APP_COMMIT=${COMMIT_SHA}
ENV APP_BUILD_DATE=${BUILD_DATE}

# Usuário não-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/main .

USER appuser

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

CMD ["./main"]
