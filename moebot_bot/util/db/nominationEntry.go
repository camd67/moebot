package db

type NominationEntry struct {
	ID   int
	Key  string
	Text string
}

func NominationOptionsQuery(nominationID int) ([]*NominationEntry, error) {
	return []*NominationEntry{}, nil
}
