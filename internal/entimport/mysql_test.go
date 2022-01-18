package entimport_test

import (
	"bytes"
	"context"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"

	"ariga.io/atlas/sql/schema"
	"ariga.io/entimport/internal/entimport"

	"entgo.io/ent/dialect"
	"github.com/go-openapi/inflect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestMySQL(t *testing.T) {
	var (
		r          = require.New(t)
		ctx        = context.Background()
		testSchema = "test"
	)
	tests := []struct {
		name           string
		entities       []string
		expectedFields map[string]string
		mock           *schema.Schema
		expectedEdges  map[string]string
	}{
		{
			name: "single_table_fields",
			mock: MockMySQLSingleTableFields(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int8("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				`user`: `func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "fields_with_attributes",
			mock: MockMySQLTableFieldsWithAttributes(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id").Comment("some id"), field.Int8("age").Optional(), field.String("name").Comment("first name"), field.String("last_name").Optional().Comment("family name")}
}`,
			},
			expectedEdges: map[string]string{
				`user`: `func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "fields_with_unique_indexes",
			mock: MockMySQLTableFieldsWithUniqueIndexes(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int8("age").Unique(), field.String("last_name").Optional().Comment("not so boring"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				`user`: `func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "multi_table_fields",
			mock: MockMySQLMultiTableFields(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int8("age").Unique(), field.String("last_name").Optional().Comment("not so boring"), field.String("name")}
}`,
				"pet": `func (Pet) Fields() []ent.Field {
	return []ent.Field{field.Int("id").Comment("pet id"), field.Int8("age").Optional(), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				`user`: `func (User) Edges() []ent.Edge {
	return nil
}`,
				`pet`: `func (Pet) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user", "pet"},
		},
		{
			name: "non_default_primary_key",
			mock: MockMySQLNonDefaultPrimaryKey(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.String("id").StorageKey("name"), field.String("last_name").Unique()}
}`,
			},
			expectedEdges: map[string]string{
				`user`: `func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_m2m_two_types",
			mock: MockMySQLM2MTwoTypes(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
				"group": `func (Group) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.From("groups", Group.Type).Ref("users")}
}`,
				"group": `func (Group) Edges() []ent.Edge {
	return []ent.Edge{edge.To("users", User.Type)}
}`,
			},
			entities: []string{"user", "group"},
		},
		{
			name: "relation_m2m_same_type",
			mock: MockMySQLM2MSameType(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type)}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_m2m_bidirectional",
			mock: MockMySQLM2MBidirectional(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type)}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_o2o_two_types",
			mock: MockMySQLO2OTwoTypes(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
				"card": `func (Card) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("number"), field.Int("user_card").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("card", Card.Type).Unique()}
}`,
				"card": `func (Card) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("card").Unique().Field("user_card")}
}`,
			},
			entities: []string{"user", "card"},
		},
		{
			name: "relation_o2o_same_type",
			mock: MockMySQLO2OSameType(),
			expectedFields: map[string]string{
				"node": `func (Node) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("value"), field.Int("node_next").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"node": `func (Node) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_node", Node.Type).Unique(), edge.From("parent_node", Node.Type).Unique().Field("node_next")}
}`,
			},
			entities: []string{"node"},
		},
		{
			name: "relation_o2o_bidirectional",
			mock: MockMySQLO2OBidirectional(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name"), field.Int("user_spouse").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_user", User.Type).Unique(), edge.From("parent_user", User.Type).Unique().Field("user_spouse")}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_o2m_two_types",
			mock: MockMySQLO2MTwoTypes(),
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
				"pet": `func (Pet) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("user_pets").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("pets", Pet.Type)}
}`,
				"pet": `func (Pet) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("pets").Unique().Field("user_pets")}
}`,
			},
			entities: []string{"user", "pet"},
		},
		{
			name: "relation_o2m_same_type",
			mock: MockMySQLO2MSameType(),
			expectedFields: map[string]string{
				"node": `func (Node) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("value"), field.Int("node_children").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"node": `func (Node) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_nodes", Node.Type), edge.From("parent_node", Node.Type).Unique().Field("node_children")}
}`,
			},
			entities: []string{"node"},
		},
		{
			name: "relation_o2x_other_side_ignored",
			mock: MockMySQLO2XOtherSideIgnored(),
			expectedFields: map[string]string{
				"pet": `func (Pet) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("user_pets").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"pet": `func (Pet) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"pet"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mockMux(ctx, dialect.MySQL, tt.mock, testSchema)
			drv, err := m.OpenImport("mysql://root:pass@tcp(localhost:3308)/test?parseTime=True")
			r.NoError(err)
			importer, err := entimport.NewImport(
				entimport.WithDriver(drv),
			)
			r.NoError(err)
			schemas := createTempDir(t)
			mutations, err := importer.SchemaMutations(ctx)
			r.NoError(err)
			err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(schemas))
			r.NoError(err)
			actualFiles := readDir(t, schemas)
			r.EqualValues(len(tt.entities), len(actualFiles))
			for _, e := range tt.entities {
				f, err := parser.ParseFile(token.NewFileSet(), "", actualFiles[e+".go"], 0)
				r.NoError(err)
				typeName := inflect.Camelize(e)
				fieldMethod := lookupMethod(f, typeName, "Fields")
				r.NotNil(fieldMethod)
				var actualFields bytes.Buffer
				err = printer.Fprint(&actualFields, token.NewFileSet(), fieldMethod)
				r.NoError(err)
				r.EqualValues(tt.expectedFields[e], actualFields.String())
				edgeMethod := lookupMethod(f, typeName, "Edges")
				r.NotNil(edgeMethod)
				var actualEdges bytes.Buffer
				err = printer.Fprint(&actualEdges, token.NewFileSet(), edgeMethod)
				r.NoError(err)
				r.EqualValues(tt.expectedEdges[e], actualEdges.String())
			}
		})
	}
}

func TestMySQLJoinTableOnly(t *testing.T) {
	var (
		testSchema = "test"
		ctx        = context.Background()
	)
	m := mockMux(ctx, dialect.MySQL, MockMySQLM2MJoinTableOnly(), testSchema)
	drv, err := m.OpenImport("mysql://root:pass@tcp(localhost:3308)/test?parseTime=True")
	require.NoError(t, err)
	importer, err := entimport.NewImport(
		entimport.WithDriver(drv),
	)
	require.NoError(t, err)
	mutations, err := importer.SchemaMutations(ctx)
	require.Empty(t, mutations)
	require.EqualError(t, err, "entimport: join tables must be inspected with ref tables - append `tables` flag")
}
