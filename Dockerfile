ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /ftpd-server ./cmd/server


FROM debian:bookworm

COPY --from=builder /ftpd-server /usr/local/bin/
ENV FTPD_PORT=8080
ENV RELAY_PORT=8443
EXPOSE $FTPD_PORT
EXPOSE $RELAY_PORT
