package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type AutoIDMixin struct {
	mixin.Schema
	CommonEntMixin
}

func (AutoIDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("ent_id", uuid.UUID{}).
			Unique().
			Default(uuid.New),
	}
}

func (AutoIDMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ent_id").Unique(),
	}
}
