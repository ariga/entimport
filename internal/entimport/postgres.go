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
	Options *ImportOptions
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
	return schemaMutations(p, s.Tables)
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
	applyColumnAttributes(f, column)
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

// smallserial- 2 bytes - small autoincrementing integer 1 to 32767
// serial - 4 bytes autoincrementing integer 1 to 2147483647
// bigserial - 8 bytes large autoincrementing integer	1 to 9223372036854775807
func (p *Postgres) convertSerial(typ *postgres.SerialType, name string) ent.Field {
	return field.Uint(name).
		SchemaType(map[string]string{
			dialect.Postgres: typ.T, // Override Postgres.
		})
}
