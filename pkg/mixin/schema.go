package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type Schema struct {
	mixin.Schema
}

type Policy interface {
	Policy() ent.Policy
}

func (Schema) Fields() []ent.Field {
	incrementalEnabled := true
	return []ent.Field{
		field.
			Uint32("auto_id").
			Unique().
			Annotations(
				entsql.Annotation{
					Incremental: &incrementalEnabled,
				}),
		field.
			Uint32("created_at").
			DefaultFunc(func() uint32 {
				return uint32(time.Now().Unix())
			}),
		field.
			Uint32("updated_at").
			DefaultFunc(func() uint32 {
				return uint32(time.Now().Unix())
			}).
			UpdateDefault(func() uint32 {
				return uint32(time.Now().Unix())
			}),
		field.
			Uint32("deleted_at").
			DefaultFunc(func() uint32 {
				return 0
			}),
	}
}
