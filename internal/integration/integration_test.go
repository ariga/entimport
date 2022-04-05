package integration

import (
	"bytes"
	"context"
	"database/sql"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"ariga.io/entimport/internal/entimport"
	"ariga.io/entimport/internal/mux"

	"entgo.io/ent/dialect"
	"github.com/go-openapi/inflect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestMySQL(t *testing.T) {
	var (
		r   = require.New(t)
		ctx = context.Background()
		dsn = "root:pass@tcp(localhost:3306)/test?parseTime=True&multiStatements=true"
	)
	var tests = []struct {
		name           string
		query          string
		entities       []string
		expectedFields map[string]string
		expectedEdges  map[string]string
	}{
		{
			name: "one table",
			// language=MySQL
			query: `
create table users
(
    id   bigint auto_increment primary key,
    age  bigint       not null,
    name varchar(255) not null
);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "int8 and int16 field types",
			// language=MySQL
			query: `
create table field_type_small_int
(
    id              bigint auto_increment primary key,
    int_8           tinyint           not null,
    int16           smallint          not null,
    optional_int8   tinyint           null,
    optional_int16  smallint          null,
    nillable_int8   tinyint           null,
    nillable_int16  smallint          null,
    optional_uint8  tinyint unsigned  null,
    optional_uint16 smallint unsigned null
);
			`,
			expectedFields: map[string]string{
				"field_type_small_int": `func (FieldTypeSmallInt) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int8("int_8"), field.Int16("int16"), field.Int8("optional_int8").Optional(), field.Int16("optional_int16").Optional(), field.Int8("nillable_int8").Optional(), field.Int16("nillable_int16").Optional(), field.Uint8("optional_uint8").Optional(), field.Uint16("optional_uint16").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"field_type_small_int": `func (FieldTypeSmallInt) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"field_type_small_int"},
		},
		{
			name: "int32 and int64 field types",
			// language=MySQL
			query: `
create table field_type_int
(
    id                      bigint auto_increment primary key,
    int_field               bigint          not null,
    int32                   int             not null,
    int64                   bigint          not null,
    optional_int            bigint          null,
    optional_int32          int             null,
    optional_int64          bigint          null,
    nillable_int            bigint          null,
    nillable_int32          int             null,
    nillable_int64          bigint          null,
    validate_optional_int32 int             null,
    optional_uint           bigint unsigned null,
    optional_uint32         int unsigned    null,
    optional_uint64         bigint unsigned null
);
			`,
			expectedFields: map[string]string{
				"field_type_int": `func (FieldTypeInt) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("int_field"), field.Int32("int32"), field.Int("int64"), field.Int("optional_int").Optional(), field.Int32("optional_int32").Optional(), field.Int("optional_int64").Optional(), field.Int("nillable_int").Optional(), field.Int32("nillable_int32").Optional(), field.Int("nillable_int64").Optional(), field.Int32("validate_optional_int32").Optional(), field.Uint64("optional_uint").Optional(), field.Uint32("optional_uint32").Optional(), field.Uint64("optional_uint64").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"field_type_int": `func (FieldTypeInt) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"field_type_int"},
		},
		{
			name: "float field types",
			// language=MySQL
			query: `create table field_type_float
(
    id               bigint auto_increment primary key,
    float_field      float   not null,
    optional_float   float   null,
    double_field     double  not null,
    optional_double  double   null,
    decimal_field    decimal not null,
    optional_decimal decimal null
);
			`,
			expectedFields: map[string]string{
				"field_type_float": `func (FieldTypeFloat) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Float32("float_field"), field.Float32("optional_float").Optional(), field.Float("double_field"), field.Float("optional_double").Optional(), field.Float("decimal_field"), field.Float("optional_decimal").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"field_type_float": `func (FieldTypeFloat) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"field_type_float"},
		},
		{
			name: "enum field types",
			// language=MySQL
			query: `
create table field_type_enum
(
    id                 bigint auto_increment primary key,
    enum_field         enum ('on', 'off')                                              null,
    enum_field_default enum ('ADMIN', 'OWNER', 'USER', 'READ', 'WRITE') default 'READ' not null
);
			`,
			expectedFields: map[string]string{
				"field_type_enum": `func (FieldTypeEnum) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Enum("enum_field").Optional().Values("on", "off"), field.Enum("enum_field_default").Values("ADMIN", "OWNER", "USER", "READ", "WRITE")}
}`,
			},
			expectedEdges: map[string]string{
				"field_type_enum": `func (FieldTypeEnum) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"field_type_enum"},
		},
		{
			name: "other field types",
			// language=MySQL
			query: `
create table field_type_other
(
    id              bigint auto_increment primary key,
    datetime        datetime     null,
    string          varchar(255) null,
    optional_string varchar(255) not null,
    bool            tinyint(1)   null,
    optional_bool   tinyint(1)   not null,
    ts              timestamp    null
);
			`,
			expectedFields: map[string]string{
				"field_type_other": `func (FieldTypeOther) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Time("datetime").Optional(), field.String("string").Optional(), field.String("optional_string"), field.Bool("bool").Optional(), field.Bool("optional_bool"), field.Time("ts").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"field_type_other": `func (FieldTypeOther) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"field_type_other"},
		},
		{
			name: "o2o two types",
			// language=MySQL
			query: `
create table users
(
    id   bigint auto_increment primary key,
    name varchar(255) not null
);

create table cards
(
    id          bigint auto_increment primary key,
    create_time timestamp not null,
    user_card   bigint    null,
    constraint user_card unique (user_card),
    constraint cards_users_card foreign key (user_card) references users (id) on delete set null
);

create index card_id on cards (id);
			`,
			entities: []string{"user", "card"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
				"card": `func (Card) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Time("create_time"), field.Int("user_card").Optional().Unique()}
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
		},
		{
			name: "o2o same type",
			// language=MySQL
			query: `
create table nodes
(
    id        bigint auto_increment primary key,
    value     bigint null,
    node_next bigint null,
    constraint node_next unique (node_next),
    constraint nodes_nodes_next foreign key (node_next) references nodes (id) on delete set null
);
			`,
			expectedFields: map[string]string{
				"node": `func (Node) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("value").Optional(), field.Int("node_next").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"node": `func (Node) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_node", Node.Type).Unique(), edge.From("parent_node", Node.Type).Ref("child_node").Unique().Field("node_next")}
}`,
			},
			entities: []string{"node"},
		},
		{
			name: "o2o bidirectional",
			// language=MySQL
			query: `
create table users
(
    id          bigint auto_increment primary key,
    name        varchar(255) not null,
    nickname    varchar(255) null,
    user_spouse bigint       null,
    constraint nickname unique (nickname),
    constraint user_spouse unique (user_spouse),
    constraint users_users_spouse foreign key (user_spouse) references users (id) on delete set null
);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.String("nickname").Optional().Unique(), field.Int("user_spouse").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_user", User.Type).Unique(), edge.From("parent_user", User.Type).Ref("child_user").Unique().Field("user_spouse")}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "o2m two types",
			// language=MySQL
			query: `
create table users
(
    id   bigint auto_increment primary key,
    name varchar(255) not null
);

create table pet
(
    id        bigint auto_increment primary key,
    name      varchar(255) not null,
    user_pets bigint       null,
    constraint pet_users_pets foreign key (user_pets) references users (id) on delete set null
);

create index pet_name_user_pets on pet (name, user_pets);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
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
			name: "o2m same type",
			// language=MySQL
			query: `
create table users
(
    id          bigint auto_increment
        primary key,
    name        varchar(255) not null,
    user_parent bigint       null,
    constraint users_users_parent foreign key (user_parent) references users (id) on delete set null
);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("user_parent").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_user", User.Type).Ref("child_users").Unique().Field("user_parent")}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "m2m bidirectional",
			// language=MySQL
			query: `
create table users
(
    id   bigint auto_increment
        primary key,
    age  bigint       not null,
    name varchar(255) not null
);

create table user_friends
(
    user_id   bigint not null,
    friend_id bigint not null,
    primary key (user_id, friend_id),
    constraint user_friends_friend_id foreign key (friend_id) references users (id) on delete cascade,
    constraint user_friends_user_id foreign key (user_id) references users (id) on delete cascade
);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type).Ref("child_users")}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "m2m same type",
			// language=MySQL
			query: `
create table users
(
    id   bigint auto_increment primary key,
    name varchar(255) not null
);

create table user_following
(
    user_id     bigint not null,
    follower_id bigint not null,
    primary key (user_id, follower_id),
    constraint user_following_follower_id foreign key (follower_id) references users (id) on delete cascade,
    constraint user_following_user_id foreign key (user_id) references users (id) on delete cascade
);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type).Ref("child_users")}
}`,
			},
			entities: []string{"user"},
		},
		{
			// Demonstrate M2M relation between two different types. User and groups.
			name: "m2m two types",
			// language=MySQL
			query: `
create table some_groups
(
    id     bigint auto_increment primary key,
    active tinyint(1) default 1 not null,
    name   varchar(255)         not null
);

create table users
(
    id   bigint auto_increment primary key,
    name varchar(255) not null
);

create table user_groups
(
    user_id  bigint not null,
    group_id bigint not null,
    primary key (user_id, group_id),
    constraint user_groups_some_groups_id foreign key (group_id) references some_groups (id) on delete cascade,
    constraint user_groups_user_id foreign key (user_id) references users (id) on delete cascade
);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
				"some_group": `func (SomeGroup) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Bool("active"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.From("some_groups", SomeGroup.Type).Ref("users")}
}`,
				"some_group": `func (SomeGroup) Edges() []ent.Edge {
	return []ent.Edge{edge.To("users", User.Type)}
}`,
			},
			entities: []string{"user", "some_group"},
		},
		{
			name: "multiple relations",
			// language=MySQL
			query: `
create table group_infos
(
    id          bigint auto_increment primary key,
    description varchar(255)         not null,
    max_users   bigint default 10000 not null
);

create table some_groups
(
    id         bigint auto_increment primary key,
    name       varchar(255) not null,
    group_info bigint       null,
    constraint groups_group_infos_info foreign key (group_info) references group_infos (id) on delete set null
);

create table users
(
    id            bigint auto_increment primary key,
    optional_int  bigint       null,
    name          varchar(255) not null,
    group_blocked bigint       null,
    constraint users_some_groups_blocked foreign key (group_blocked) references some_groups (id) on delete set null
);

create table user_groups
(
    user_id  bigint not null,
    group_id bigint not null,
    primary key (user_id, group_id),
    constraint user_groups_some_groups_id foreign key (group_id) references some_groups (id) on delete cascade,
    constraint user_groups_user_id foreign key (user_id) references users (id) on delete cascade
);
			`,
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("optional_int").Optional(), field.String("name"), field.Int("group_blocked").Optional()}
}`,
				"group_info": `func (GroupInfo) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("description"), field.Int("max_users")}
}`,
				"some_group": `func (SomeGroup) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("group_info_id").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.From("some_groups", SomeGroup.Type).Ref("users"), edge.From("some_group", SomeGroup.Type).Ref("users").Unique().Field("group_blocked")}
}`,
				"group_info": `func (GroupInfo) Edges() []ent.Edge {
	return []ent.Edge{edge.To("some_groups", SomeGroup.Type)}
}`,
				"some_group": `func (SomeGroup) Edges() []ent.Edge {
	return []ent.Edge{edge.From("group_info", GroupInfo.Type).Ref("some_groups").Unique().Field("group_info_id"), edge.To("users", User.Type), edge.To("users", User.Type)}
}`,
			},
			entities: []string{"user", "group_info", "some_group"},
		},
	}
	db, err := sql.Open(dialect.MySQL, dsn)
	r.NoError(err)
	defer db.Close()
	r.NoError(db.Ping())
	drv, err := mux.Default.OpenImport("mysql://" + dsn)
	r.NoError(err)
	defer drv.Close()
	si, err := entimport.NewImport(
		entimport.WithDriver(drv),
	)
	r.NoError(err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			dropMySQL(t, db)
			schemas := createTempDir(t)
			_, err := db.ExecContext(ctx, tt.query)
			r.NoError(err)
			mutations, err := si.SchemaMutations(ctx)
			r.NoError(err)
			err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(schemas))
			r.NoError(err)
			r.NotZero(tt.entities)
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

func TestPostgres(t *testing.T) {
	var (
		r   = require.New(t)
		ctx = context.Background()
		dsn = "postgres://postgres:pass@localhost:5432/test?sslmode=disable"
	)
	tests := []struct {
		name           string
		query          string
		entities       []string
		expectedFields map[string]string
		expectedEdges  map[string]string
	}{
		{
			name: "one table",
			// language=PostgreSQL
			query: `
create table users
(
    id   bigint primary key,
    age  bigint       not null,
    name varchar(255) not null
)
			`,
			entities: []string{"user"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return nil
}`,
			},
		},
		{
			name: "field types - int8 and int16",
			// language=PostgreSQL
			query: `
create table field_types
(
    id             bigint generated by default as identity
        constraint field_types_pkey
            primary key,
    int8           smallint not null,
    int16          smallint not null,
    optional_int8  smallint,
    optional_int16 smallint
);
		`,
			entities: []string{"field_type"},
			expectedFields: map[string]string{
				"field_type": `func (FieldType) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int16("int8"), field.Int16("int16"), field.Int16("optional_int8").Optional(), field.Int16("optional_int16").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"field_type": `func (FieldType) Edges() []ent.Edge {
	return nil
}`,
			},
		},
		{
			name: "int32 and int64 field types",
			// language=PostgreSQL
			query: `
create table field_types
(
    id             bigint generated by default as identity
        constraint field_types_pkey
            primary key,
    int            bigint  not null,
    int32          integer not null,
    int64          bigint  not null,
    optional_int   bigint,
    optional_int32 integer,
    optional_int64 bigint
);
		`,
			entities: []string{"field_type"},
			expectedFields: map[string]string{
				"field_type": `func (FieldType) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("int"), field.Int32("int32"), field.Int("int64"), field.Int("optional_int").Optional(), field.Int32("optional_int32").Optional(), field.Int("optional_int64").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"field_type": `func (FieldType) Edges() []ent.Edge {
	return nil
}`,
			},
		},
		{
			name: "float field types",
			// language=PostgreSQL
			query: `
create table field_types
(
    id               bigint generated by default as identity
        constraint field_types_pkey
            primary key,
    optional_float64   double precision,
    optional_float32 real
);

		`,
			entities: []string{"field_type"},
			expectedFields: map[string]string{
				"field_type": `func (FieldType) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Float("optional_float64").Optional(), field.Float32("optional_float32").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"field_type": `func (FieldType) Edges() []ent.Edge {
	return nil
}`,
			},
		},
		{
			name: "other field types",
			// language=PostgreSQL
			query: `
create table field_types
(
    id              bigint generated by default as identity
        constraint field_types_pkey
            primary key,
    datetime        date,
    decimal         numeric,
    string          varchar                                    not null,
    optional_string varchar,
    bool            boolean,
    ts              timestamp with time zone,
    string_default  varchar default 'READ':: character varying not null
);
		`,
			entities: []string{"field_type"},
			expectedFields: map[string]string{
				"field_type": `func (FieldType) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Time("datetime").Optional(), field.Float("decimal").Optional(), field.String("string"), field.String("optional_string").Optional(), field.Bool("bool").Optional(), field.Time("ts").Optional(), field.String("string_default")}
}`,
			},
			expectedEdges: map[string]string{
				"field_type": `func (FieldType) Edges() []ent.Edge {
	return nil
}`,
			},
		},
		{
			// See https://entgo.io/docs/schema-edges#relationship.
			name: "o2o two types",
			// language=PostgreSQL
			query: `
create table users
(
    id           bigint generated by default as identity
        constraint users_pkey primary key,
    optional_int bigint,
    name         varchar                                    not null,
    nickname     varchar
        constraint users_nickname_key unique,
    role         varchar default 'user':: character varying not null
);

create table cards
(
    id          bigint generated by default as identity
        constraint cards_pkey
            primary key,
    create_time timestamp with time zone   not null,
    balance     double precision default 0 not null,
    name        varchar,
    user_card   bigint
        constraint cards_user_card_key
            unique
        constraint cards_users_card
            references users
            on delete set null
);

create index card_id on cards (id);
`,
			entities: []string{"user", "card"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("optional_int").Optional(), field.String("name"), field.String("nickname").Optional().Unique(), field.String("role")}
}`,
				"card": `func (Card) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Time("create_time"), field.Float("balance"), field.String("name").Optional(), field.Int("user_card").Optional().Unique()}
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
		},
		{
			name: "o2o same type",
			// language=PostgreSQL
			query: `
create table nodes
(
    id        bigint generated by default as identity
        constraint nodes_pkey primary key,
    value     bigint,
    node_next bigint
        constraint nodes_node_next_key unique
        constraint nodes_nodes_next
            references nodes
            on delete set null
);
					`,
			entities: []string{"node"},
			expectedFields: map[string]string{
				"node": `func (Node) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("value").Optional(), field.Int("node_next").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"node": `func (Node) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_node", Node.Type).Unique(), edge.From("parent_node", Node.Type).Ref("child_node").Unique().Field("node_next")}
}`,
			},
		},
		{
			name: "o2o bidirectional",
			// language=PostgreSQL
			query: `
create table users
(
    id          bigint generated by default as identity
        constraint users_pkey
            primary key,
    name        varchar not null,
    user_spouse bigint
        constraint users_user_spouse_key
            unique
        constraint users_users_spouse
            references users
            on delete set null
);

					`,
			entities: []string{"user"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("user_spouse").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_user", User.Type).Unique(), edge.From("parent_user", User.Type).Ref("child_user").Unique().Field("user_spouse")}
}`,
			},
		},
		{
			name: "o2m two types",
			// language=PostgreSQL
			query: `
create table users
(
    id   bigint generated by default as identity
        constraint users_pkey primary key,
    name varchar not null
);

create table pet
(
    id        bigint generated by default as identity
        constraint pet_pkey
            primary key,
    name      varchar not null,
    user_pets bigint
        constraint pet_users_pets
            references users
            on delete set null
);

create index pet_name_user_pets
    on pet (name, user_pets);
					`,
			entities: []string{"user", "pet"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
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
		},
		{
			name: "o2m same type",
			// language=PostgreSQL
			query: `
create table users
(
    id          bigint generated by default as identity
        constraint users_pkey
            primary key,
    name        varchar not null,
    user_parent bigint
        constraint users_users_parent
            references users
            on delete set null
);
					`,
			entities: []string{"user"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("user_parent").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_user", User.Type).Ref("child_users").Unique().Field("user_parent")}
}`,
			},
		},
		{
			name: "m2m bidirectional",
			// language=PostgreSQL
			query: `
create table users
(
    id   bigint generated by default as identity
        constraint users_pkey
            primary key,
    name varchar not null
);

create table user_friends
(
    user_id   bigint not null
        constraint user_friends_user_id
            references users
            on delete cascade,
    friend_id bigint not null
        constraint user_friends_friend_id
            references users
            on delete cascade,
    constraint user_friends_pkey
        primary key (user_id, friend_id)
);

					`,
			entities: []string{"user"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type).Ref("child_users")}
}`,
			},
		},
		{
			name: "m2m same type",
			// language=PostgreSQL
			query: `
create table users
(
    id   bigint generated by default as identity
        constraint users_pkey
            primary key,
    name varchar                                      not null,
    last varchar default 'unknown'::character varying not null
);

create table user_following
(
    user_id     bigint not null
        constraint user_following_user_id
            references users
            on delete cascade,
    follower_id bigint not null
        constraint user_following_follower_id
            references users
            on delete cascade,
    constraint user_following_pkey
        primary key (user_id, follower_id)
);
					`,
			entities: []string{"user"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.String("last")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type).Ref("child_users")}
}`,
			},
		},
		{
			name: "m2m two types",
			// language=PostgreSQL
			query: `
create table users
(
    id   bigint generated by default as identity
        constraint users_pkey
            primary key,
    name varchar not null
);

create table groups
(
    id     bigint generated by default as identity
        constraint groups_pkey
            primary key,
    active boolean default true not null,
    name   varchar              not null
);

create table user_groups
(
    user_id  bigint not null
        constraint user_groups_user_id
            references users
            on delete cascade,
    group_id bigint not null
        constraint user_groups_group_id
            references groups
            on delete cascade,
    constraint user_groups_pkey
        primary key (user_id, group_id)
);
					`,
			entities: []string{"user", "group"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
				"group": `func (Group) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Bool("active"), field.String("name")}
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
		},
		{
			name: "multiple relations",
			// language=PostgreSQL
			query: `
create table group_infos
(
    id          bigint generated by default as identity
        constraint group_infos_pkey primary key,
    description varchar              not null,
    max_users   bigint default 10000 not null
);

create table users
(
    id   bigint generated by default as identity
        constraint users_pkey
            primary key,
    name varchar not null
);

create table groups
(
    id         bigint generated by default as identity
        constraint groups_pkey
            primary key,
    name       varchar not null,
    group_info bigint
        constraint groups_group_infos_info
            references group_infos
            on delete set null
);

create table user_groups
(
    user_id  bigint not null
        constraint user_groups_user_id
            references users
            on delete cascade,
    group_id bigint not null
        constraint user_groups_group_id
            references groups
            on delete cascade,
    constraint user_groups_pkey
        primary key (user_id, group_id)
);
					`,
			entities: []string{"user", "group", "group_info"},
			expectedFields: map[string]string{
				"user": `func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
				"group": `func (Group) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("group_info_id").Optional()}
}`,
				"group_info": `func (GroupInfo) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("description"), field.Int("max_users")}
}`,
			},
			expectedEdges: map[string]string{
				"user": `func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.From("groups", Group.Type).Ref("users")}
}`,
				"group": `func (Group) Edges() []ent.Edge {
	return []ent.Edge{edge.From("group_info", GroupInfo.Type).Ref("groups").Unique().Field("group_info_id"), edge.To("users", User.Type)}
}`,
				"group_info": `func (GroupInfo) Edges() []ent.Edge {
	return []ent.Edge{edge.To("groups", Group.Type)}
}`,
			},
		},
	}

	db, err := sql.Open(dialect.Postgres, dsn)
	r.NoError(err)
	defer db.Close()
	r.NoError(db.Ping())
	drv, err := mux.Default.OpenImport(dsn)
	r.NoError(err)
	defer drv.Close()
	si, err := entimport.NewImport(
		entimport.WithDriver(drv),
	)
	r.NoError(err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			dropPostgres(t, db)
			schemas := createTempDir(t)
			_, err := db.ExecContext(ctx, tt.query)
			r.NoError(err)
			mutations, err := si.SchemaMutations(ctx)
			r.NoError(err)
			err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(schemas))
			r.NoError(err)
			r.NotZero(tt.entities)
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

func createTempDir(t *testing.T) string {
	r := require.New(t)
	tmpDir, err := ioutil.TempDir("", "entimport-*")
	r.NoError(err)
	t.Cleanup(func() {
		err = os.RemoveAll(tmpDir)
		r.NoError(err)
	})
	return tmpDir
}

func readDir(t *testing.T, path string) map[string]string {
	files := make(map[string]string)
	err := filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}
		buf, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.Base(path)] = string(buf)
		return nil
	})
	require.NoError(t, err)
	return files
}

func dropMySQL(t *testing.T, db *sql.DB) {
	r := require.New(t)
	t.Log("drop data from database")
	ctx := context.Background()
	_, err := db.ExecContext(ctx, "DROP DATABASE IF EXISTS test")
	r.NoError(err)
	_, err = db.ExecContext(ctx, "CREATE DATABASE test")
	r.NoError(err)
	_, _ = db.ExecContext(ctx, "USE test")
}

func dropPostgres(t *testing.T, db *sql.DB) {
	r := require.New(t)
	t.Log("drop data from database")
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`)
	r.NoError(err)
}

func lookupMethod(file *ast.File, typeName string, methodName string) (m *ast.FuncDecl) {
	ast.Inspect(file, func(node ast.Node) bool {
		if decl, ok := node.(*ast.FuncDecl); ok {
			if decl.Name.Name != methodName || decl.Recv == nil || len(decl.Recv.List) != 1 {
				return true
			}
			if id, ok := decl.Recv.List[0].Type.(*ast.Ident); ok && id.Name == typeName {
				m = decl
				return false
			}
		}
		return true
	})
	return m
}
