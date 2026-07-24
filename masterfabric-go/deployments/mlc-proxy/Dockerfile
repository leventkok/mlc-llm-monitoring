FROM golang:1.26.4-alpine AS builder
WORKDIR /src
COPY main.go .
RUN go mod init local/mlc-proxy && CGO_ENABLED=0 go build -o /bin/mlc-proxy .

FROM alpine:3.24
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /bin/mlc-proxy /usr/local/bin/mlc-proxy
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mlc-proxy"]
