package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type CommonEntMixin struct {
	mixin.Schema
}

func (CommonEntMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("id"),
	}
}
