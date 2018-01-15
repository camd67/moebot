FROM golang

COPY / /go/src/github.com/camd67/www

RUN go install github.com/camd67/www

CMD ["/go/bin/www"]

ENV port 8040
EXPOSE 8040
