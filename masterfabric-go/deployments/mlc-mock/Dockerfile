FROM golang:1.26.4-alpine AS builder
WORKDIR /src
COPY main.go .
RUN go mod init local/mlc-mock && CGO_ENABLED=0 go build -o /bin/mlc-mock .

FROM alpine:3.24
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /bin/mlc-mock /usr/local/bin/mlc-mock
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mlc-mock"]
