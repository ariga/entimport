package entimport

import (
	"context"
	"errors"
	"fmt"

	"ariga.io/atlas/sql/schema"
	"ariga.io/entimport/internal/mux"

	"entgo.io/contrib/schemast"
	"entgo.io/ent"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/dialect/entsql"
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

	// relOptions are the options passed down to the functions that create a relation.
	relOptions struct {
		uniqueEdgeToChild    bool
		recursive            bool
		uniqueEdgeFromParent bool
		refName              string
		edgeField            string
	}

	// fieldFunc receives an Atlas column and converts it to an Ent field.
	fieldFunc func(column *schema.Column) (f ent.Field, err error)

	// SchemaImporter is the interface that wraps the SchemaMutations method.
	SchemaImporter interface {
		// SchemaMutations imports a given schema from a data source and returns a list of schemast mutators.
		SchemaMutations(context.Context) ([]schemast.Mutator, error)
	}

	// ImportOptions are the options passed on to every SchemaImporter.
	ImportOptions struct {
		tables     []string
		schemaPath string
		driver     *mux.ImportDriver
	}

	// ImportOption allows for managing import configuration using functional options.
	ImportOption func(*ImportOptions)
)

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

// WithDriver provides an import driver to be used by SchemaImporter.
func WithDriver(drv *mux.ImportDriver) ImportOption {
	return func(i *ImportOptions) {
		i.driver = drv
	}
}

// NewImport calls the relevant data source importer based on a given dialect.
func NewImport(opts ...ImportOption) (SchemaImporter, error) {
	var (
		si  SchemaImporter
		err error
	)
	i := &ImportOptions{}
	for _, apply := range opts {
		apply(i)
	}
	switch i.driver.Dialect {
	case dialect.MySQL:
		si, err = NewMySQL(i)
		if err != nil {
			return nil, err
		}
	case dialect.Postgres:
		si, err = NewPostgreSQL(i)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("entimport: unsupported dialect %q", i.driver.Dialect)
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

// entEdge creates an edge based on the given params and direction.
func entEdge(nodeName, nodeType string, currentNode *schemast.UpsertSchema, dir edgeDir, opts relOptions) (e ent.Edge) {
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

// setEdgeField is a function to properly name edge fields.
func setEdgeField(e ent.Edge, opts relOptions, childNode *schemast.UpsertSchema) {
	edgeField := opts.edgeField
	// rename the field in case the edge and the field have the same name
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

// upsertRelation takes 2 nodes and created the edges between them.
func upsertRelation(nodeA *schemast.UpsertSchema, nodeB *schemast.UpsertSchema, opts relOptions) {
	tableA := tableName(nodeA.Name)
	tableB := tableName(nodeB.Name)
	opts.refName = tableB
	fromA := entEdge(tableA, nodeA.Name, nodeB, from, opts)
	toB := entEdge(tableB, nodeB.Name, nodeA, to, opts)
	nodeA.Edges = append(nodeA.Edges, toB)
	nodeB.Edges = append(nodeB.Edges, fromA)
}

// upsertManyToMany handles the creation of M2M relations.
func upsertManyToMany(mutations map[string]schemast.Mutator, table *schema.Table) error {
	tableA := table.ForeignKeys[0].RefTable
	tableB := table.ForeignKeys[1].RefTable
	var opts relOptions
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

// resolvePrimaryKey returns the primary key as an ent field for a given table.
func resolvePrimaryKey(field fieldFunc, table *schema.Table) (f ent.Field, err error) {
	if table.PrimaryKey == nil || len(table.PrimaryKey.Parts) != 1 {
		return nil, fmt.Errorf("entimport: invalid primary key - single part key must be present")
	}
	if f, err = field(table.PrimaryKey.Parts[0].C); err != nil {
		return nil, err
	}
	if d := f.Descriptor(); d.Name != "id" {
		d.StorageKey = d.Name
		d.Name = "id"
	}
	return f, nil
}

// upsertNode handles the creation of a node from a given table.
func upsertNode(field fieldFunc, table *schema.Table) (*schemast.UpsertSchema, error) {
	upsert := &schemast.UpsertSchema{
		Name: typeName(table.Name),
	}
	fields := make(map[string]ent.Field, len(upsert.Fields))
	for _, f := range upsert.Fields {
		fields[f.Descriptor().Name] = f
	}
	pk, err := resolvePrimaryKey(field, table)
	if err != nil {
		return nil, err
	}
	if _, ok := fields[pk.Descriptor().Name]; !ok {
		fields[pk.Descriptor().Name] = pk
		upsert.Fields = append(upsert.Fields, pk)
	}
	for _, column := range table.Columns {
		if table.PrimaryKey != nil &&
			len(table.PrimaryKey.Parts) != 0 &&
			table.PrimaryKey.Parts[0].C.Name == column.Name {
			continue
		}
		fld, err := field(column)
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
			fld, ok := fields[column.Name]
			if !ok {
				return nil, fmt.Errorf("foreign key for column: %q doesn't exist in referenced table", column.Name)
			}
			fld.Descriptor().Optional = true
		}
	}
	return upsert, err
}

// applyColumnAttributes adds column attributes to a given ent field.
func applyColumnAttributes(f ent.Field, col *schema.Column) {
	desc := f.Descriptor()
	desc.Optional = col.Type.Null
	for _, attr := range col.Attrs {
		if a, ok := attr.(*schema.Comment); ok {
			desc.Comment = a.Text
		}
	}
}

// schemaMutations is in charge of creating all the schema mutations needed for an ent schema.
func schemaMutations(field fieldFunc, tables []*schema.Table) ([]schemast.Mutator, error) {
	mutations := make(map[string]schemast.Mutator)
	joinTables := make(map[string]*schema.Table)
	for _, table := range tables {
		if isJoinTable(table) {
			joinTables[table.Name] = table
			continue
		}
		node, err := upsertNode(field, table)
		if err != nil {
			return nil, err
		}
		node.Annotations = []entschema.Annotation{entsql.Annotation{Table: table.Name}}
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
	idxs := make(map[string]*schema.Index)
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
		opts := relOptions{
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
