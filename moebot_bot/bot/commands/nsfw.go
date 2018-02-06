package commands

type NsfwCommand struct{}

func (nc NsfwCommand) Execute(pack *CommPackage) {
	// force NSFW comm param so we can reuse guild role
	processGuildRole([]string{"NSFW"}, pack.Session, []string{"NSFW"}, pack.Channel, pack.Guild, pack.Message, false)
}

func (nc NsfwCommand) Setup() {}
