FROM golang:1.15

WORKDIR /go/src/ferda
COPY . .

RUN go get -v github.com/bwmarrin/discordgo
RUN go get -v github.com/jmoiron/sqlx
RUN go get -v github.com/lib/pq

CMD ["go", "run", "main.go"]