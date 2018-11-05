package types

type GroupType int

const (
	// Group type where any role can be selected, and multiple can be selected
	GroupTypeAny = 1
	// Group type where only one role can be selected
	GroupTypeExclusive = 2
	// Same as the exclusive group, but can't be removed
	GroupTypeExclusiveNoRemove = 3
	// Same as the exclusive group, but will prevent changing if one of the roles is present
	GroupTypeNoMultiples = 4

	OptionsForGroupType = "ANY, EXC, ENR, NOM"
)

type RoleGroup struct {
	Id       int
	ServerId int
	Name     string
	Type     GroupType
}
