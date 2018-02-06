package commands

type TeamCommand struct{}

func (tc TeamCommand) Execute(pack *CommPackage) {
	processGuildRole([]string{"Nanachi", "Ozen", "Bondrewd", "Reg", "Riko", "Maruruk"}, pack.Session, pack.Params, pack.Channel, pack.Guild, pack.Message, false)
}
