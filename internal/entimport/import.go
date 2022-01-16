package entimport

import (
	"context"
	"errors"
	"fmt"

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
	// field receives an Atlas column and converts in to an ent field.
	field(column *schema.Column) (f ent.Field, err error)
}

type ImportOptions struct {
	dsn        string
	tables     []string
	schemaPath string
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
		edgeField += "_id"
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

func typeName(tableName string) string {
	return inflect.Camelize(inflect.Singularize(tableName))
}

func tableName(typeName string) string {
	return inflect.Underscore(inflect.Pluralize(typeName))
}

func resolvePrimaryKey(importer SchemaImporter, table *schema.Table) (f ent.Field, err error) {
	if table.PrimaryKey == nil || len(table.PrimaryKey.Parts) != 1 {
		return nil, fmt.Errorf("entimport: invalid primary key - single part key must be present")
	}
	if f, err = importer.field(table.PrimaryKey.Parts[0].C); err != nil {
		return nil, err
	}
	if f.Descriptor().Name != "id" {
		f.Descriptor().StorageKey = f.Descriptor().Name
		f.Descriptor().Name = "id"
	}
	return f, nil
}

func upsertNode(i SchemaImporter, table *schema.Table) (*schemast.UpsertSchema, error) {
	upsert := &schemast.UpsertSchema{
		Name: typeName(table.Name),
	}
	fields := make(map[string]ent.Field, len(upsert.Fields))
	for _, f := range upsert.Fields {
		fields[f.Descriptor().Name] = f
	}
	pk, err := resolvePrimaryKey(i, table)
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
		fld, err := i.field(column)
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
	return upsert, err
}

func applyColumnAttributes(f ent.Field, col *schema.Column) {
	desc := f.Descriptor()
	desc.Optional = col.Type.Null
	for _, attr := range col.Attrs {
		if a, ok := attr.(*schema.Comment); ok {
			desc.Comment = a.Text
		}
	}
}

func schemaMutations(importer SchemaImporter, tables []*schema.Table) ([]schemast.Mutator, error) {
	mutations := make(map[string]schemast.Mutator, len(tables))
	joinTables := make(map[string]*schema.Table)
	for _, table := range tables {
		if isJoinTable(table) {
			joinTables[table.Name] = table
			continue
		}
		node, err := upsertNode(importer, table)
		if err != nil {
			return nil, err
		}
		mutations[table.Name] = node
	}
	for _, table := range tables {
		if t, ok := joinTables[table.Name]; ok {
			err := upsertManyToMany(mutations, t)
			if err != nil {
				return nil, err
			}
			continue
		}
		upsertOneToX(mutations, table)
	}
	ml := make([]schemast.Mutator, 0, len(mutations))
	for _, mutator := range mutations {
		ml = append(ml, mutator)
	}
	return ml, nil
}

// O2O Two Types - Child Table has a unique reference (FK) to Parent table
// O2O Same Type - Child Table has a unique reference (FK) to Parent table (itself)
// O2M (The "Many" side, keeps a reference to the "One" side).
// O2M Two Types - Parent has a non-unique reference to Child, and Child has a unique back-reference to Parent
// O2M Same Type - Parent has a non-unique reference to Child, and Child doesn't have a back-reference to Parent.
func upsertOneToX(mutations map[string]schemast.Mutator, table *schema.Table) {
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
		colName := fk.Columns[0].Name
		opts := options{
			uniqueEdgeFromParent: true,
			refName:              child.Name,
			edgeField:            colName,
		}
		if child.Name == parent.Name {
			opts.recursive = true
		}
		idx, ok := idxs[colName]
		if ok && idx.Unique {
			opts.uniqueEdgeToChild = true
		}
		// If at least one table in the relation does not exist, there is no point to create it.
		parentNode, ok := mutations[parent.Name].(*schemast.UpsertSchema)
		if !ok {
			return
		}
		childNode, ok := mutations[child.Name].(*schemast.UpsertSchema)
		if !ok {
			return
		}
		upsertRelation(parentNode, childNode, opts)
	}
}
