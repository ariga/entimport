package entimport

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/schema"

	"entgo.io/contrib/schemast"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"github.com/go-openapi/inflect"
	mysqldriver "github.com/go-sql-driver/mysql"
)

const (
	mTinyInt   = "tinyint"   // MYSQL_TYPE_TINY
	mSmallInt  = "smallint"  // MYSQL_TYPE_SHORT
	mInt       = "int"       // MYSQL_TYPE_LONG
	mMediumInt = "mediumint" // MYSQL_TYPE_INT24
	mBigInt    = "bigint"    // MYSQL_TYPE_LONGLONG
)

// MySQL holds the schema import options and an Atlas inspector instance
type MySQL struct {
	schema.Inspector
	Options   *ImportOptions
	mutations map[string]schemast.Mutator
}

// NewMySQL - create aמ import structure for MySQL.
func NewMySQL(opts ...ImportOption) (*MySQL, error) {
	i := &ImportOptions{}
	for _, apply := range opts {
		apply(i)
	}
	db, err := sql.Open(dialect.MySQL, i.dsn)
	if err != nil {
		return nil, fmt.Errorf("entimport: failed to open db connection: %w", err)
	}
	drv, err := mysql.Open(db)
	if err != nil {
		return nil, fmt.Errorf("entimport: error while trying to open db inspection client %w", err)
	}
	return &MySQL{
		Inspector: drv,
		Options:   i,
	}, nil
}

// SchemaMutations implements SchemaImporter.
func (m *MySQL) SchemaMutations(ctx context.Context) ([]schemast.Mutator, error) {
	inspectOptions := &schema.InspectOptions{
		Tables: m.Options.tables,
	}
	// dsn example: root:pass@tcp(localhost:3308)/test?parseTime=True
	cfg, err := mysqldriver.ParseDSN(m.Options.dsn)
	if err != nil {
		return nil, err
	}
	if cfg.DBName == "" {
		return nil, errors.New("DSN connection string must include schema(database) name")
	}
	s, err := m.Inspector.InspectSchema(ctx, cfg.DBName, inspectOptions)
	if err != nil {
		return nil, err
	}
	m.mutations = make(map[string]schemast.Mutator, len(s.Tables))
	return m.schemaMutations(s.Tables)
}

func (m *MySQL) schemaMutations(tables []*schema.Table) ([]schemast.Mutator, error) {
	joinTables := make(map[string]*schema.Table)
	for _, table := range tables {
		if isJoinTable(table) {
			joinTables[table.Name] = table
			continue
		}
		if _, err := m.upsertNode(table); err != nil {
			return nil, err
		}
	}
	for _, table := range tables {
		if t, ok := joinTables[table.Name]; ok {
			err := upsertManyToMany(m.mutations, t)
			if err != nil {
				return nil, err
			}
			continue
		}
		m.upsertOneToX(table)
	}
	ml := make([]schemast.Mutator, 0, len(m.mutations))
	for _, mutator := range m.mutations {
		ml = append(ml, mutator)
	}
	return ml, nil
}

func (m *MySQL) resolvePrimaryKey(table *schema.Table) (f ent.Field, err error) {
	if table.PrimaryKey == nil || len(table.PrimaryKey.Parts) != 1 {
		return nil, fmt.Errorf("entimport: invalid primary key - single part key must be present")
	}
	if f, err = m.field(table.PrimaryKey.Parts[0].C); err != nil {
		return nil, err
	}
	if f.Descriptor().Name != "id" {
		f.Descriptor().StorageKey = f.Descriptor().Name
		f.Descriptor().Name = "id"
	}
	return f, nil
}

func (m *MySQL) field(column *schema.Column) (f ent.Field, err error) {
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
		f = m.convertFloat(typ, name)
	case *schema.IntegerType:
		f = m.convertInteger(typ, name)
	case *schema.JSONType:
		f = field.JSON(name, json.RawMessage{})
	case *schema.StringType:
		f = field.String(name)
	case *schema.TimeType:
		f = field.Time(name)
	default:
		return nil, fmt.Errorf("entimport: unsupported type %q", typ)
	}
	m.applyColumnAttributes(f, column)
	return f, err
}

func (m *MySQL) convertFloat(typ *schema.FloatType, name string) (f ent.Field) {
	// A precision from 0 to 23 results in a 4-byte single-precision FLOAT column.
	// A precision from 24 to 53 results in an 8-byte double-precision DOUBLE column:
	// https://dev.mysql.com/doc/refman/8.0/en/floating-point-types.html
	if typ.Precision > 23 {
		return field.Float(name)
	}
	return field.Float32(name)
}

func (m *MySQL) convertInteger(typ *schema.IntegerType, name string) (f ent.Field) {
	if typ.Unsigned {
		switch typ.T {
		case mTinyInt:
			f = field.Uint8(name)
		case mSmallInt:
			f = field.Uint16(name)
		case mMediumInt:
			f = field.Uint32(name)
		case mInt:
			f = field.Uint32(name)
		case mBigInt:
			f = field.Uint64(name)
		}
		return f
	}
	switch typ.T {
	case mTinyInt:
		f = field.Int8(name)
	case mSmallInt:
		f = field.Int16(name)
	case mMediumInt:
		f = field.Int32(name)
	case mInt:
		f = field.Int32(name)
	case mBigInt:
		// Int64 is not used on purpose.
		f = field.Int(name)
	}
	return f
}

func (m *MySQL) applyColumnAttributes(f ent.Field, col *schema.Column) {
	desc := f.Descriptor()
	desc.Optional = col.Type.Null
	for _, attr := range col.Attrs {
		if a, ok := attr.(*schema.Comment); ok {
			desc.Comment = a.Text
		}
	}
	for _, idx := range col.Indexes {
		if idx.Unique && len(idx.Parts) == 1 {
			desc.Unique = idx.Unique
		}
	}
	// FK / Reference column
	if col.ForeignKeys != nil {
		desc.Optional = true
	}
}

// O2O Two Types - Child Table has a unique reference (FK) to Parent table
// O2O Same Type - Child Table has a unique reference (FK) to Parent table (itself)
// O2M (The "Many" side, keeps a reference to the "One" side).
// O2M Two Types - Parent has a non-unique reference to Child, and Child has a unique back-reference to Parent
// O2M Same Type - Parent has a non-unique reference to Child, and Child doesn't have a back-reference to Parent.
func (m *MySQL) upsertOneToX(table *schema.Table) {
	if table.ForeignKeys == nil || table.Indexes == nil {
		return
	}
	for _, fk := range table.ForeignKeys {
		if len(fk.Columns) != 1 {
			continue
		}
		for _, idx := range table.Indexes {
			if len(idx.Parts) != 1 {
				continue
			}
			// MySQL requires indexes on foreign keys and referenced keys.
			if fk.Columns[0] == idx.Parts[0].C {
				parent := fk.RefTable
				child := table
				opts := options{
					uniqueEdgeFromParent: true,
					refName:              child.Name,
				}
				if child.Name == parent.Name {
					opts.recursive = true
				}
				if idx.Unique {
					opts.uniqueEdgeToChild = true
				}
				opts.edgeField = fk.Columns[0].Name
				// If at least one table in the relation does not exist, there is no point to create it.
				parentNode, ok := m.mutations[parent.Name].(*schemast.UpsertSchema)
				if !ok {
					return
				}
				childNode, ok := m.mutations[child.Name].(*schemast.UpsertSchema)
				if !ok {
					return
				}
				upsertRelation(parentNode, childNode, opts)
			}
		}
	}
}

func (m *MySQL) upsertNode(table *schema.Table) (*schemast.UpsertSchema, error) {
	upsert := &schemast.UpsertSchema{
		Name: typeName(table.Name),
	}
	if u, ok := m.mutations[table.Name].(*schemast.UpsertSchema); ok {
		upsert = u
	}
	fields := make(map[string]ent.Field, len(upsert.Fields))
	for _, f := range upsert.Fields {
		fields[f.Descriptor().Name] = f
	}
	pk, err := m.resolvePrimaryKey(table)
	if err != nil {
		return nil, err
	}
	if _, ok := fields[pk.Descriptor().Name]; !ok {
		upsert.Fields = append(upsert.Fields, pk)
	}
	for _, column := range table.Columns {
		if column.Name == table.PrimaryKey.Parts[0].C.Name {
			continue
		}
		fld, err := m.field(column)
		if err != nil {
			return nil, err
		}
		if _, ok := fields[column.Name]; !ok {
			upsert.Fields = append(upsert.Fields, fld)
		}
	}
	m.mutations[table.Name] = upsert
	return upsert, err
}

func typeName(tableName string) string {
	return inflect.Camelize(inflect.Singularize(tableName))
}

func tableName(typeName string) string {
	return inflect.Underscore(inflect.Pluralize(typeName))
}
