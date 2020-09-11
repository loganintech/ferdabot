FROM golang:1.15
\
WORKDIR /go/src/ferda
COPY . .

RUN go get -v github.com/bwmarrin/discordgo
RUN go get -v github.com/jmoiron/sqlx
RUN go get -v github.com/lib/pq

ENV TZ 'America/Los_Angeles'
RUN echo $TZ > /etc/timezone && \
    apt-get update && apt-get install -y tzdata && \
    rm /etc/localtime && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure -f noninteractive tzdata && \
    apt-get clean

CMD ["go", "run", "src/main.go"]