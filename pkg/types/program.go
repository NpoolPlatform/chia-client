package types

type Program struct {
}

// TODO
type Condition interface {
	ToProgram() Program
	FromProgram(program Program) Condition
}

type MyCondition struct {
}

func (e *MyCondition) ToProgram() Program {
	return Program{}
}

func (e *MyCondition) FromProgram(program Program) Condition {
	return &MyCondition{}
}
