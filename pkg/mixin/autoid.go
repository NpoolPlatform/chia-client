package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type AutoIDMixin struct {
	mixin.Schema
}

func (AutoIDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("id"),
		field.UUID("ent_id", uuid.UUID{}).
			Unique().
			Default(uuid.New),
	}
}
