package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type AutoIDMixin struct {
	mixin.Schema
}

func (AutoIDMixin) Fields() []ent.Field {
	enabled := true
	return []ent.Field{
		field.
			UUID("id", uuid.UUID{}).
			Unique().
			Default(uuid.New),
		field.
			Uint32("auto_id").
			Unique().
			Annotations(entsql.Annotation{
				Incremental: &enabled,
			}),
	}
}
