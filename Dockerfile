FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM scratch

COPY --from=builder /api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
