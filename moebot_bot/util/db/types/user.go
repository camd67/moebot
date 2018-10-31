package types

type UserProfile struct {
	Id      int
	UserUid string
}

type UserServerRank struct {
	Id          int
	ServerId    int
	UserId      int
	Rank        int
	MessageSent bool
}

type UserServerRankWrapper struct {
	UserUid   string
	ServerUid string
	Rank      int
	SendTo    string
}
