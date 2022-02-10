FROM golang:1.17 AS builder

WORKDIR /go/src/github.com/loganintech/ferdabot
COPY . .

RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ferdabot .

FROM debian:buster

RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates

RUN /usr/sbin/update-ca-certificates

WORKDIR /ferda/
COPY --from=builder /go/src/github.com/loganintech/ferdabot/ferdabot .
CMD ["./ferdabot"]