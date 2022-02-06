package databuilder

import "k8s.io/apimachinery/pkg/util/sets"

type plan struct {
	order    []*builder
	initData sets.String
}

func (p *plan) Run(_ ...interface{}) (Data, error) {
	panic("not implemented") // TODO: Implement
}

func newPlan(order []*builder, initData []string) (Plan, error) {
	return &plan{
		order:    order,
		initData: sets.NewString(initData...),
	}, nil
}
