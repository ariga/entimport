package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"ariga.io/entimport/internal/entimport"
)

var tablesFlag tables

func init() {
	flag.Var(&tablesFlag, "tables", "comma-separated list of tables to inspect (all if empty)")
}

func main() {
	dsn := flag.String("dsn", "", "data source name (connection information)")
	schemaPath := flag.String("schema-path", "./ent/schema", "output path for ent schema")
	dialect := flag.String("dialect", "mysql", "database dialect")
	flag.Parse()
	if *dsn == "" {
		log.Println("entimport: data source name (dsn) must be provided")
		flag.Usage()
		os.Exit(2)
	}
	ctx := context.Background()
	i, err := entimport.NewImport(*dialect,
		entimport.WithDSN(*dsn),
		entimport.WithTables(tablesFlag))
	if err != nil {
		log.Fatalf("entimport: create importer (%s) failed - %v", *dialect, err)
	}
	mutations, err := i.SchemaMutations(ctx)
	if err != nil {
		log.Fatalf("entimport: schema import failed - %v", err)
	}
	if err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(*schemaPath)); err != nil {
		log.Fatalf("entimport: schema writing failed - %v", err)
	}
}

type tables []string

func (t *tables) String() string {
	return fmt.Sprint(*t)
}

func (t *tables) Set(s string) error {
	*t = strings.Split(s, ",")
	return nil
}
