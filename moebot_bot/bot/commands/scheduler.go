package commands

type Scheduler interface {
	Execute(operationID int64)
	Keyword() string
	AddScheduledOperation(comm *CommPackage) error
	Help() string
	OperationDescription(operationID int64) string
}
