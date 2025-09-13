package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/edge"
)

// School holds the schema definition for the School entity.
type School struct {
	ent.Schema
}

// Fields of the School.
func (School) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique(),
		field.String("address").
			NotEmpty(),
		field.String("city").
			NotEmpty(),
		field.String("state").
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the School.
func (School) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("courses", Course.Type), // A school can have multiple courses
	}
}
