// Code generated by entimport, DO NOT EDIT.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Pet struct {
	ent.Schema
}

func (Pet) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name").Optional(), field.Int8("sex").Optional(), field.Time("create_time").Optional()}
}
func (Pet) Edges() []ent.Edge {
	return nil
}
func (Pet) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "pet"}}
}