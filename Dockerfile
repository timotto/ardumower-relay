# This Dockerfile is not used by the CI to build the releases.
# Use this Dockerfile to build your own Docker images without the CI.
FROM golang:1.17 AS build

ADD . /build
WORKDIR /build
RUN go test -race ./...
RUN CGO_ENABLED=0 go build -o /relay -a -tags netgo -ldflags "-w" ./cmd/relay

FROM scratch
COPY --from=build /relay /relay
VOLUME /config
ENTRYPOINT ["/relay", "/config/config.yml"]
