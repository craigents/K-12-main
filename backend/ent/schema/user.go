package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/edge"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			Unique().
			NotEmpty(),
		field.String("hashed_password").
			NotEmpty().
			Sensitive(), // Mark as sensitive if you don't want it in default String() output
		field.String("first_name").
			Default(""),
		field.String("last_name").
			Default(""),
		field.String("email").
			Unique().
			NotEmpty(),
		field.Enum("role").
			Values("student", "teacher", "admin").
			Default("student"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("taught_courses", Course.Type).
			From("teacher"), // This will be defined more clearly when Course schema is detailed
		edge.To("enrolled_courses", Course.Type).
			From("students"), // This will be defined more clearly when Course schema is detailed
	}
}
