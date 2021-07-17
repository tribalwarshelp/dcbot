package model

import "github.com/go-pg/pg/v10/orm"

type DefaultFilter struct {
	Limit  int
	Offset int
	Order  []string
}

func (f *DefaultFilter) Apply(q *orm.Query) (*orm.Query, error) {
	if f.Limit != 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset != 0 {
		q = q.Offset(f.Offset)
	}
	if len(f.Order) > 0 {
		q = q.Order(f.Order...)
	}
	return q, nil
}
