FROM golang

COPY . /go/src/github.com/camd67/moebot/api

RUN go install github.com/camd67/moebot/api

CMD ["/go/bin/api"]

ENV port 9090
EXPOSE 9090
