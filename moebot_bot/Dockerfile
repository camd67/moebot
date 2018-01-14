FROM golang

# copy from current directory to output dir
COPY / /go/src/github.com/camd67/moebot/moebot_bot

# this comand is pretty slow and error prone... should figure out a better way to copy deps
RUN go get "github.com/bwmarrin/discordgo" "github.com/lib/pq" "golang.org/x/image/font" "golang.org/x/image/font/basicfont" "golang.org/x/image/math/fixed"
RUN go install "github.com/camd67/moebot/moebot_bot"

CMD ["/go/bin/moebot_bot"]

EXPOSE 80
