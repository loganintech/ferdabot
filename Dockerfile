FROM golang:1.15 AS builder

WORKDIR /go/src/github.com/loganintech/ferdabot
COPY . .

RUN go get -v github.com/bwmarrin/discordgo
RUN go get -v github.com/jmoiron/sqlx
RUN go get -v github.com/lib/pq

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ferdabot .

FROM debian:buster

ADD ./cert.crt /usr/local/share/ca-certificates/cert.crt
RUN chmod 644 /usr/local/share/ca-certificates/cert.crt
RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates

RUN /usr/sbin/update-ca-certificates

WORKDIR /ferda/
COPY --from=builder /go/src/github.com/loganintech/ferdabot/ferdabot .
CMD ["./ferdabot"]