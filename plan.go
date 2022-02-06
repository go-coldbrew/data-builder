package databuilder

type plan struct {
	order []*builder
}

func (p *plan) Run(_ ...interface{}) (Data, error) {
	panic("not implemented") // TODO: Implement
}
