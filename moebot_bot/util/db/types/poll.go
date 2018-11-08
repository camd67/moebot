package types

type PollOption struct {
	Id           int
	PollId       int
	ReactionId   string
	ReactionName string
	Description  string
	Votes        int
}

type Poll struct {
	Id         int
	Options    []*PollOption
	Title      string
	Open       bool
	ChannelId  int
	UserUid    string
	MessageUid string
}
