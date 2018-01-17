FROM golang

# copy from current directory to output dir
COPY . /go/src/github.com/camd67/moebot/moebot_bot

RUN go install github.com/camd67/moebot/moebot_bot

CMD ["/go/bin/moebot_bot"]

EXPOSE 80