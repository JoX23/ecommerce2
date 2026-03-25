FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
# config.yaml NO se copia en la imagen. En producción la config se inyecta
# exclusivamente via variables de entorno (APP_* prefix).
EXPOSE 8080
CMD ["./server"]
