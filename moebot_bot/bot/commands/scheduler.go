package commands

type Scheduler interface {
	Execute(operationID int)
	Keyword() string
	AddScheduledOperation(comm *CommPackage)
}
