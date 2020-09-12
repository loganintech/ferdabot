FROM golang:1.15 AS builder

WORKDIR /go/src/github.com/loganintech/ferdabot
COPY . .

RUN go get -v github.com/bwmarrin/discordgo && \
    go get -v github.com/jmoiron/sqlx && \
    go get -v github.com/lib/pq

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ferdabot .

FROM alpine:latest
WORKDIR /ferda/
COPY --from=builder /go/src/github.com/loganintech/ferdabot/ferdabot .
CMD ["./ferdabot"]