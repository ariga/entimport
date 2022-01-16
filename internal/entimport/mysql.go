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
	Options *ImportOptions
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
	return schemaMutations(m, s.Tables)
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
	applyColumnAttributes(f, column)
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
