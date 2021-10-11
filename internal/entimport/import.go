package entimport

import (
	"context"
	"errors"
	"fmt"
	"io"

	"ariga.io/atlas/sql/schema"

	"entgo.io/contrib/schemast"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"github.com/go-openapi/inflect"
)

const (
	header         = "Code generated " + "by entimport, DO NOT EDIT."
	to     edgeDir = iota
	from
)

var joinTableErr = errors.New("entimport: join tables must be inspected with ref tables - append `tables` flag")

type (
	edgeDir int
	options struct {
		uniqueEdgeToChild    bool
		recursive            bool
		uniqueEdgeFromParent bool
		refName              string
		edgeField            string
	}
)

// ImportOption allows for managing import configuration using functional options.
type ImportOption func(*ImportOptions)

// WithDSN provides a DSN (data source name) for reading the schema & tables from.
// DSN must include a schema (named-database) in the connection string.
func WithDSN(dsn string) ImportOption {
	return func(i *ImportOptions) {
		i.dsn = dsn
	}
}

// WithSchemaPath provides a DSN (data source name) for reading the schema & tables from.
func WithSchemaPath(path string) ImportOption {
	return func(i *ImportOptions) {
		i.schemaPath = path
	}
}

// WithTables limits the schema import to a set of given tables (by all tables are imported)
func WithTables(tables []string) ImportOption {
	return func(i *ImportOptions) {
		i.tables = tables
	}
}

// SchemaImporter is the interface that wraps the SchemaMutations method.
type SchemaImporter interface {
	// SchemaMutations imports a given schema from a data source and returns a list of schemast mutators.
	SchemaMutations(context.Context) ([]schemast.Mutator, error)
}

type ImportOptions struct {
	dsn         string
	tables      []string
	schemaPath  string
	importEdges bool
	writer      io.Writer
}

// NewImport - calls the relevant data source importer based on a given dialect.
func NewImport(dialectName string, opts ...ImportOption) (SchemaImporter, error) {
	var (
		si  SchemaImporter
		err error
	)
	switch dialectName {
	case dialect.MySQL:
		si, err = NewMySQL(opts...)
		if err != nil {
			return nil, err
		}
	case dialect.Postgres:
		si, err = NewPostgreSQL(opts...)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("entimport: unsupported dialect %q", dialectName)
	}
	return si, err
}

// WriteSchema receives a list of mutators, and writes an ent schema to a given location in the file system.
func WriteSchema(mutations []schemast.Mutator, opts ...ImportOption) error {
	i := &ImportOptions{}
	for _, apply := range opts {
		apply(i)
	}
	ctx, err := schemast.Load(i.schemaPath)
	if err != nil {
		return err
	}
	if err = schemast.Mutate(ctx, mutations...); err != nil {
		return err
	}
	return ctx.Print(i.schemaPath, schemast.Header(header))
}

func entEdge(nodeName, nodeType string, currentNode *schemast.UpsertSchema, dir edgeDir, opts options) (e ent.Edge) {
	switch dir {
	case to:
		e = edge.To(nodeName, ent.Schema.Type)
		if opts.uniqueEdgeToChild {
			e.Descriptor().Unique = true
			e.Descriptor().Name = inflect.Singularize(nodeName)
		}
		if opts.recursive {
			e.Descriptor().Name = "child_" + e.Descriptor().Name
		}
	case from:
		e = edge.From(nodeName, ent.Schema.Type)
		if opts.uniqueEdgeFromParent {
			e.Descriptor().Unique = true
			e.Descriptor().Name = inflect.Singularize(nodeName)
		}
		if opts.edgeField != "" {
			setEdgeField(e, opts, currentNode)
		}
		if opts.recursive {
			e.Descriptor().Name = "parent_" + e.Descriptor().Name
			break
		}
		// RefName describes which entEdge of the Parent Node we're referencing
		// because there can be multiple references from one node to another.
		refName := opts.refName
		if opts.uniqueEdgeToChild {
			refName = inflect.Singularize(refName)
		}
		e.Descriptor().RefName = refName
	}
	e.Descriptor().Type = nodeType
	return e
}

func setEdgeField(e ent.Edge, opts options, childNode *schemast.UpsertSchema) {
	edgeField := opts.edgeField
	if e.Descriptor().Name == edgeField {
		edgeField = edgeField + "_id"
		for _, f := range childNode.Fields {
			if f.Descriptor().Name == opts.edgeField {
				f.Descriptor().Name = edgeField
			}
		}
	}
	e.Descriptor().Field = edgeField
}

func upsertRelation(nodeA *schemast.UpsertSchema, nodeB *schemast.UpsertSchema, opts options) {
	tableA := tableName(nodeA.Name)
	tableB := tableName(nodeB.Name)
	opts.refName = tableB
	fromA := entEdge(tableA, nodeA.Name, nodeB, from, opts)
	toB := entEdge(tableB, nodeB.Name, nodeA, to, opts)
	nodeA.Edges = append(nodeA.Edges, toB)
	nodeB.Edges = append(nodeB.Edges, fromA)
}

func upsertManyToMany(mutations map[string]schemast.Mutator, table *schema.Table) error {
	tableA := table.ForeignKeys[0].RefTable
	tableB := table.ForeignKeys[1].RefTable
	var opts options
	if tableA.Name == tableB.Name {
		opts.recursive = true
	}
	nodeA, ok := mutations[tableA.Name].(*schemast.UpsertSchema)
	if !ok {
		return joinTableErr
	}
	nodeB, ok := mutations[tableB.Name].(*schemast.UpsertSchema)
	if !ok {
		return joinTableErr
	}
	upsertRelation(nodeA, nodeB, opts)
	return nil
}

// Note: at this moment ent doesn't support fields on m2m relations.
func isJoinTable(table *schema.Table) bool {
	if table.PrimaryKey == nil || len(table.PrimaryKey.Parts) != 2 || len(table.ForeignKeys) != 2 {
		return false
	}
	// Make sure that the foreign key columns exactly match primary key column.
	for _, fk := range table.ForeignKeys {
		if len(fk.Columns) != 1 {
			return false
		}
		if fk.Columns[0] != table.PrimaryKey.Parts[0].C && fk.Columns[0] != table.PrimaryKey.Parts[1].C {
			return false
		}
	}
	return true
}
