package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/edge"
)

// Course holds the schema definition for the Course entity.
type Course struct {
	ent.Schema
}

// Fields of the Course.
func (Course) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.String("code").
			Unique().
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("grade_level").
			Optional(), // E.g., "9", "10", "K"
		field.String("subject").
			Optional(), // E.g., "Math", "Science"
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Course.
func (Course) Edges() []ent.Edge {
	return []ent.Edge{
		// A course belongs to one school.
		edge.From("school", School.Type).
			Ref("courses"). // References the "courses" edge in the School schema
			Unique().       // A course must have one school
			Required(),     // This field is required

		// A course can have multiple teachers.
		edge.From("teachers", User.Type).
			Ref("taught_courses"), // References the "taught_courses" edge in User schema

		// A course can have multiple enrolled students.
		edge.From("students", User.Type).
			Ref("enrolled_courses"), // References the "enrolled_courses" edge in User schema
	}
}
