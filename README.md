# entimport

`entimport` is a tool for creating [Ent](https://entgo.io/) schemas from existing SQL databases. Currently, `MySQL`
and `PostgreSQL` are supported. The tool can import to [ent schema](https://entgo.io/docs/schema-def) any number of
tables, including relations between them.

## Installation

### Setup A Go Environment

If your project directory is outside [GOPATH](https://github.com/golang/go/wiki/GOPATH) or you are not familiar with
GOPATH, setup a [Go module](https://github.com/golang/go/wiki/Modules#quick-start) project as follows:

```shell
go mod init <project>
```

### Install ent

```shell
go install entgo.io/ent/cmd/ent
```

After installing `ent` codegen tool, you should have it in your `PATH`. If you don't find it your path, you can also
run: `go run entgo.io/ent/cmd/ent <command>`

### Create Schema Directory

Go to the root directory of your project, and run:

```shell
ent init
```

The command above will create `<project>/ent/schema/` directory and the file inside `<project>/ent/generate.go`

### Importing a Schema

Installing and running `entimport`

```shell
go run ariga.io/entimport/cmd/entimport
```

- For example, importing a MySQL schema with `users` table:

```shell
go run ariga.io/entimport/cmd/entimport -dsn "mysql://root:pass@tcp(localhost:3308)/test" -tables "users"
```

The command above will write a valid ent schema into the directory specified (or the default `./ent/schema`):

```
.
├── generate.go
└── schema
    └── user.go

1 directory, 2 files
```

### Code Generation:

In order to [generate](https://entgo.io/docs/code-gen) `ent` files from the produced schemas, run:

```shell
go run -mod=mod entgo.io/ent/cmd/ent generate ./schema

# OR `ent` init:

go generate ./ent
```

If you are not yet familiar with `ent`, you can also follow
the [quick start guide](https://entgo.io/docs/getting-started).

## Usage

```shell
entimport  -h
```

```
Usage of ./entimport:
  -dsn string
        data source name (connection information), for example:
        "mysql://user:pass@tcp(localhost:3306)/dbname"
        "postgres://user:pass@host:port/dbname"
  -schema-path string
        output path for ent schema (default "./ent/schema")
  -tables value
        comma-separated list of tables to inspect (all if empty)
```

## Examples:

1. Import ent schema from Postgres database

> Note: add search_path=foo if you use non `public` schema.

```shell
go run ariga.io/entimport/cmd/entimport -dsn "postgres://postgres:pass@localhost:5432/test?sslmode=disable" 
```

2. Import ent schema from MySQL database

```shell
go run ariga.io/entimport/cmd/entimport -dsn "mysql://root:pass@tcp(localhost:3308)/test"
```

3. Import only specific tables:

> Note: When importing specific tables:  
> if the table is a join table, you must also provide referenced tables.  
> If the table is only one part of a relation, the other part won't be imported unless specified.   
> If the `-tables` flags is omitted all tables in current `database schema` will be imported

```shell
go run ariga.io/entimport/cmd/entimport -dsn "..." -tables "users,user_friends" 
```

4. Import to another directory:

```shell
go run ariga.io/entimport/cmd/entimport -dsn "..." -schema-path "/some/path/here"
```

## Future Work

- Index support (currently Unique index is supported).
- Support for all data types (for example `uuid` in Postgres).
- Support for Default value in columns.
- Support for editing schema both manually and automatically (real upsert and not only overwrite)
- Postgres special types: postgres.NetworkType, postgres.BitType, *schema.SpatialType, postgres.CurrencyType,
  postgres.XMLType, postgres.ArrayType, postgres.UserDefinedType.

### Known Caveats:

- Schema files are overwritten by new calls to `entimport`.
- There is no difference in DB schema between `O2O Bidirectional` and `O2O Same Type` - both will result in the same
  `ent` schema.
- There is no difference in DB schema between `M2M Bidirectional` and `M2M Same Type` - both will result in the same
 `ent` schema.
- In recursive relations the `edge` names will be prefixed with `child_` & `parent_`.
- For example: `users` with M2M relation to itself will result in:

```go
func (User) Edges() []ent.Edge {
return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type)}
}
```

## Feedback & Support

For discussion and support, [open an issue](https://github.com/ariga/entimport/issues/new/choose) or join
our [channel](https://gophers.slack.com/archives/C01FMSQDT53) in the gophers Slack.
