FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -o /mlc-agent .

FROM docker:27-cli
COPY --from=build /mlc-agent /usr/local/bin/mlc-agent
WORKDIR /deployments
ENTRYPOINT ["/usr/local/bin/mlc-agent"]
