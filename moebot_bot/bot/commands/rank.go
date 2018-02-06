package commands

type RankCommand struct{}

func (rc RankCommand) Execute(pack *CommPackage) {
	processGuildRole([]string{"Red", "Blue"}, pack.Session, pack.Params, pack.Channel, pack.Guild, pack.Message, true)
}
