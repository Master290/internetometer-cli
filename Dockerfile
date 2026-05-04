FROM golang:1.25-trixie AS builder

RUN apt update && apt -y install ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o internetometer-exporter cmd/prom/exporter.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/internetometer-exporter /
EXPOSE 9112

CMD [ "/internetometer-exporter" ]
