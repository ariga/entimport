package entimport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"

	"entgo.io/contrib/schemast"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Postgres implements SchemaImporter for PostgreSQL databases.
type Postgres struct {
	schema.Inspector
	Options   *ImportOptions
	mutations map[string]schemast.Mutator
}

// NewPostgreSQL - returns a new *Postgres.
func NewPostgreSQL(opts ...ImportOption) (*Postgres, error) {
	i := &ImportOptions{}
	for _, apply := range opts {
		apply(i)
	}
	db, err := sql.Open(dialect.Postgres, i.dsn)
	if err != nil {
		return nil, fmt.Errorf("entimport: failed to open db connection: %w", err)
	}
	drv, err := postgres.Open(db)
	if err != nil {
		return nil, fmt.Errorf("entimport: error while trying to open db inspection client %w", err)
	}
	return &Postgres{
		Inspector: drv,
		Options:   i,
	}, nil
}

// SchemaMutations implements SchemaImporter.
func (p *Postgres) SchemaMutations(ctx context.Context) ([]schemast.Mutator, error) {
	inspectOptions := &schema.InspectOptions{
		Tables: p.Options.tables,
	}
	// dsn example: "host=localhost port=5434 user=postgres dbname=test password=pass sslmode=disable search_path=public"
	schemaName := "public"
	r := regexp.MustCompile(`search_path=(\S+)`)
	matches := r.FindStringSubmatch(p.Options.dsn)
	if len(matches) != 0 {
		schemaName = matches[1]
	}

	s, err := p.Inspector.InspectSchema(ctx, schemaName, inspectOptions)
	if err != nil {
		return nil, err
	}
	p.mutations = make(map[string]schemast.Mutator, len(s.Tables))
	return p.schemaMutations(s.Tables)
}

func (p *Postgres) schemaMutations(tables []*schema.Table) ([]schemast.Mutator, error) {
	joinTables := make(map[string]*schema.Table)
	for _, table := range tables {
		if isJoinTable(table) {
			joinTables[table.Name] = table
			continue
		}
		if _, err := p.upsertNode(table); err != nil {
			return nil, err
		}
	}
	for _, table := range tables {
		if t, ok := joinTables[table.Name]; ok {
			err := upsertManyToMany(p.mutations, t)
			if err != nil {
				return nil, err
			}
			continue
		}
		p.upsertOneToX(table)
	}
	ml := make([]schemast.Mutator, 0, len(p.mutations))
	for _, mutator := range p.mutations {
		ml = append(ml, mutator)
	}
	return ml, nil
}

// O2O Two Types - Child Table has a unique reference (FK) to Parent table
// O2O Same Type - Child Table has a unique reference (FK) to Parent table (itself)
// O2M (The "Many" side, keeps a reference to the "One" side).
// O2M Two Types - Parent has a non-unique reference to Child, and Child has a unique back-reference to Parent
// O2M Same Type - Parent has a non-unique reference to Child, and Child doesn't have a back-reference to Parent.
func (p *Postgres) upsertOneToX(table *schema.Table) {
	if table.ForeignKeys == nil {
		return
	}
	idxs := make(map[string]*schema.Index, len(table.Indexes))
	for _, idx := range table.Indexes {
		if len(idx.Parts) != 1 {
			continue
		}
		idxs[idx.Parts[0].C.Name] = idx
	}
	for _, fk := range table.ForeignKeys {
		if len(fk.Columns) != 1 {
			continue
		}
		parent := fk.RefTable
		child := table
		col := fk.Columns[0].Name
		opts := options{
			uniqueEdgeFromParent: true,
			refName:              child.Name,
			edgeField:            col,
		}
		if child.Name == parent.Name {
			opts.recursive = true
		}
		idx, ok := idxs[col]
		if ok && idx.Unique {
			opts.uniqueEdgeToChild = true
		}
		// If at least one table in the relation does not exist, there is no point to create it.
		parentNode, ok := p.mutations[parent.Name].(*schemast.UpsertSchema)
		if !ok {
			return
		}
		childNode, ok := p.mutations[child.Name].(*schemast.UpsertSchema)
		if !ok {
			return
		}
		upsertRelation(parentNode, childNode, opts)
	}
}

func (p *Postgres) resolvePrimaryKey(table *schema.Table) (f ent.Field, err error) {
	if table.PrimaryKey == nil || len(table.PrimaryKey.Parts) != 1 {
		return nil, fmt.Errorf("entimport: invalid primary key - single part key must be present")
	}
	if f, err = p.field(table.PrimaryKey.Parts[0].C); err != nil {
		return nil, err
	}
	if f.Descriptor().Name != "id" {
		f.Descriptor().StorageKey = f.Descriptor().Name
		f.Descriptor().Name = "id"
	}
	return f, nil
}

func (p *Postgres) field(column *schema.Column) (f ent.Field, err error) {
	name := column.Name
	switch typ := column.Type.Type.(type) {
	case *schema.BinaryType:
		f = field.Bytes(name)
	case *schema.BoolType:
		f = field.Bool(name)
	case *schema.DecimalType:
		f = field.Float(name)
	case *schema.EnumType:
		f = field.Enum(name).Values(typ.Values...)
	case *schema.FloatType:
		f = p.convertFloat(typ, name)
	case *schema.IntegerType:
		f = p.convertInteger(typ, name)
	case *schema.JSONType:
		f = field.JSON(name, json.RawMessage{})
	case *schema.StringType:
		f = field.String(name)
	case *schema.TimeType:
		f = field.Time(name)
	case *postgres.SerialType:
		f = p.convertSerial(typ, name)
	case *postgres.UUIDType:
		f = field.UUID(name, uuid.New())
	default:
		return nil, fmt.Errorf("entimport: unsupported type %q", typ)
	}
	p.applyColumnAttributes(f, column)
	return f, err
}

// decimal, numeric - user-specified precision, exact up to 131072 digits before the decimal point;
// up to 16383 digits after the decimal point.
// real - 4 bytes variable-precision, inexact 6 decimal digits precision.
// double -	8 bytes	variable-precision, inexact	15 decimal digits precision.
func (p *Postgres) convertFloat(typ *schema.FloatType, name string) (f ent.Field) {
	if typ.Precision > 14 {
		return field.Float(name)
	}
	return field.Float32(name)
}

func (p *Postgres) convertInteger(typ *schema.IntegerType, name string) (f ent.Field) {
	switch typ.T {
	// smallint - 2 bytes small-range integer -32768 to +32767.
	case "smallint":
		f = field.Int16(name)
	// integer - 4 bytes typical choice for integer	-2147483648 to +2147483647.
	case "integer":
		f = field.Int32(name)
	// bigint - 8 bytes large-range integer	-9223372036854775808 to 9223372036854775807.
	case "bigint":
		// Int64 is not used on purpose.
		f = field.Int(name)
	}
	return f
}

func (p *Postgres) applyColumnAttributes(f ent.Field, col *schema.Column) {
	desc := f.Descriptor()
	desc.Optional = col.Type.Null
	for _, attr := range col.Attrs {
		if a, ok := attr.(*schema.Comment); ok {
			desc.Comment = a.Text
		}
	}
}

func (p *Postgres) upsertNode(table *schema.Table) (*schemast.UpsertSchema, error) {
	upsert := &schemast.UpsertSchema{
		Name: typeName(table.Name),
	}
	if u, ok := p.mutations[table.Name].(*schemast.UpsertSchema); ok {
		upsert = u
	}
	fields := make(map[string]ent.Field, len(upsert.Fields))
	for _, f := range upsert.Fields {
		fields[f.Descriptor().Name] = f
	}
	pk, err := p.resolvePrimaryKey(table)
	if err != nil {
		return nil, err
	}
	if _, ok := fields[pk.Descriptor().Name]; !ok {
		fields[pk.Descriptor().Name] = pk
		upsert.Fields = append(upsert.Fields, pk)
	}
	for _, column := range table.Columns {
		if column.Name == table.PrimaryKey.Parts[0].C.Name {
			continue
		}
		fld, err := p.field(column)
		if err != nil {
			return nil, err
		}
		if _, ok := fields[column.Name]; !ok {
			fields[column.Name] = fld
			upsert.Fields = append(upsert.Fields, fld)
		}
	}
	for _, index := range table.Indexes {
		if index.Unique && len(index.Parts) == 1 {
			fields[index.Parts[0].C.Name].Descriptor().Unique = true
		}
	}
	for _, fk := range table.ForeignKeys {
		for _, column := range fk.Columns {
			// FK / Reference column
			fields[column.Name].Descriptor().Optional = true
		}
	}
	p.mutations[table.Name] = upsert
	return upsert, err
}

// smallserial- 2 bytes - small autoincrementing integer 1 to 32767
// serial - 4 bytes autoincrementing integer 1 to 2147483647
// bigserial - 8 bytes large autoincrementing integer	1 to 9223372036854775807
func (p *Postgres) convertSerial(typ *postgres.SerialType, name string) ent.Field {
	return field.Uint(name).
		SchemaType(map[string]string{
			dialect.Postgres: typ.T, // Override Postgres.
		})
}
