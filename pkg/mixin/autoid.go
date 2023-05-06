package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type AutoIDMixin struct {
	mixin.Schema
}

func (AutoIDMixin) Fields() []ent.Field {
	incrementalEnabled := true
	return []ent.Field{
		field.
			Uint32("auto_id").
			Unique().
			Annotations(
				entsql.Annotation{
					Incremental: &incrementalEnabled,
				}),
	}
}
