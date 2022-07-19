package entimport

import (
	"context"
	"encoding/json"
	"fmt"

	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/schema"

	"entgo.io/contrib/schemast"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
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
	*ImportOptions
}

// NewMySQL - create aמ import structure for MySQL.
func NewMySQL(i *ImportOptions) (*MySQL, error) {
	return &MySQL{
		ImportOptions: i,
	}, nil
}

// SchemaMutations implements SchemaImporter.
func (m *MySQL) SchemaMutations(ctx context.Context) ([]schemast.Mutator, error) {
	inspectOptions := &schema.InspectOptions{
		Tables: m.tables,
	}
	s, err := m.driver.InspectSchema(ctx, m.driver.SchemaName, inspectOptions)
	if err != nil {
		return nil, err
	}
	var tables []*schema.Table
	if m.excludedTables != nil {
		excludedTableNames := make(map[string]bool)
		for _, t := range m.excludedTables {
			excludedTableNames[t] = true
		}
		// filter out tables that are in excludedTables:
		for _, t := range s.Tables {
			if !excludedTableNames[t.Name] {
				tables = append(tables, t)
			} else {
			}
		}
	}
	return schemaMutations(m.field, tables)
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
		return nil, fmt.Errorf("column %v: unsupported type %q", column.Name, typ)
	}
	applyColumnAttributes(f, column)
	return f, err
}

func (m *MySQL) convertFloat(typ *schema.FloatType, name string) (f ent.Field) {
	// A precision from 0 to 23 results in a 4-byte single-precision FLOAT column.
	// A precision from 24 to 53 results in an 8-byte double-precision DOUBLE column:
	// https://dev.mysql.com/doc/refman/8.0/en/floating-point-types.html
	if typ.T == mysql.TypeDouble {
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
