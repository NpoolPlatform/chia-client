package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type AuthIDMixin struct {
	mixin.Schema
}

func (AuthIDMixin) Fields() []ent.Field {
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
