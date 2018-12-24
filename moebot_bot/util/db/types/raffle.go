package types

// Raffle type enum
type RaffleType int

type RaffleEntry struct {
	Id               int
	GuildUid         string
	UserUid          string
	RaffleType       RaffleType
	TicketCount      int
	RaffleData       string
	LastTicketUpdate int64
}
