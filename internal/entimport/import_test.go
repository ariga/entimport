package entimport_test

import (
	"context"
	"go/ast"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"ariga.io/entimport/internal/mux"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func MockMySQLTableNameDoesNotUsePluralForm() *schema.Schema {
	table := &schema.Table{
		Name: "pet",
		Columns: []*schema.Column{
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "tinyint",
						Unsigned: false,
					},
					Raw:  "tinyint",
					Null: false,
				},
			},
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Default: &schema.RawExpr{X: "unknown"},
				Attrs: []schema.Attr{
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C: &schema.Column{
					Name: "id",
					Type: &schema.ColumnType{
						Type: &schema.IntegerType{
							T:        "bigint",
							Unsigned: false,
						},
						Raw:  "bigint",
						Null: false,
					},
					Attrs: []schema.Attr{
						&mysql.AutoIncrement{
							V: 0,
						},
					},
				},
			},
		},
	}
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "test",
		Tables: []*schema.Table{table},
	}
}

func MockMySQLSingleTableFields() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "tinyint",
						Unsigned: false,
					},
					Raw:  "tinyint",
					Null: false,
				},
			},
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Default: &schema.RawExpr{X: "unknown"},
				Attrs: []schema.Attr{
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C: &schema.Column{
					Name: "id",
					Type: &schema.ColumnType{
						Type: &schema.IntegerType{
							T:        "bigint",
							Unsigned: false,
						},
						Raw:  "bigint",
						Null: false,
					},
					Attrs: []schema.Attr{
						&mysql.AutoIncrement{
							V: 0,
						},
					},
				},
			},
		},
	}
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "test",
		Tables: []*schema.Table{table},
	}
}

func MockMySQLTableFieldsWithAttributes() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "tinyint",
						Unsigned: false,
					},
					Raw:  "tinyint",
					Null: true,
				},
			},
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "some id"},
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Default: &schema.RawExpr{X: "unknown"},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "first name"},
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: true,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "family name"},
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C: &schema.Column{
					Name: "id",
					Type: &schema.ColumnType{
						Type: &schema.IntegerType{
							T:        "bigint",
							Unsigned: false,
						},
						Raw:  "bigint",
						Null: false,
					},
					Attrs: []schema.Attr{
						&schema.Comment{Text: "some id"},
						&mysql.AutoIncrement{
							V: 0,
						},
					},
				},
			},
		},
	}
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "test",
		Tables: []*schema.Table{table},
	}
}

func MockMySQLTableFieldsWithUniqueIndexes() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "tinyint",
						Unsigned: false,
					},
					Raw:  "tinyint",
					Null: false,
				},
				Indexes: []*schema.Index{
					{
						Name:   "users_age_uindex",
						Unique: true,
						Attrs: []schema.Attr{
							&mysql.IndexType{
								T: "BTREE",
							},
						},
						Parts: []*schema.IndexPart{
							{
								SeqNo: 1,
								Attrs: []schema.Attr{
									&schema.Collation{V: "A"},
								},
							},
						},
					},
				},
			},
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: true,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "not so boring"},
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Default: &schema.RawExpr{X: "unknown"},
				Attrs: []schema.Attr{
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
		},
	}
	table.Indexes = []*schema.Index{
		{
			Name:   "users_age_uindex",
			Unique: true,
			Attrs: []schema.Attr{
				&mysql.IndexType{
					T: "BTREE",
				},
			},
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&schema.Collation{V: "A"},
					},
					C: table.Columns[0],
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[1],
			},
		},
	}
	return &schema.Schema{
		Name:   "test",
		Tables: []*schema.Table{table},
	}
}

func MockMySQLMultiTableFields() *schema.Schema {
	tableA := &schema.Table{Name: "users",
		Columns: []*schema.Column{
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "tinyint",
						Unsigned: false,
					},
					Raw:  "tinyint",
					Null: false,
				}, Indexes: []*schema.Index{
					{
						Name:   "users_age_uindex",
						Unique: true,
						Attrs: []schema.Attr{
							&mysql.IndexType{
								T: "BTREE",
							},
						},
						Parts: []*schema.IndexPart{
							{
								SeqNo: 1,
								Attrs: []schema.Attr{
									&schema.Collation{V: "A"},
								},
							},
						},
					},
				},
			},
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: true,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "not so boring"},
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Default: &schema.RawExpr{X: "unknown"},
				Attrs: []schema.Attr{
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
		},
	}
	tableA.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     tableA.Columns[1],
			},
		},
	}
	tableA.Indexes = []*schema.Index{
		{
			Name:   "users_age_uindex",
			Unique: true,
			Attrs: []schema.Attr{
				&mysql.IndexType{
					T: "BTREE",
				},
			},
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&schema.Collation{V: "A"},
					},
					C: tableA.Columns[0],
				},
			},
		},
	}
	tableB := &schema.Table{
		Name: "pets",
		Columns: []*schema.Column{
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "tinyint",
						Unsigned: false,
					},
					Raw:  "tinyint",
					Null: true,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Default: &schema.RawExpr{X: "unknown"},
				Attrs: []schema.Attr{
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "pet id"},
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
		},
	}
	tableB.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     tableB.Columns[2],
			},
		},
	}
	return &schema.Schema{
		Name: "test",
		Tables: []*schema.Table{
			tableA,
			tableB,
		},
	}
}

func MockMySQLNonDefaultPrimaryKey() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[1],
			},
		},
	}
	table.Indexes = []*schema.Index{
		{
			Name:   "users_last_name_uindex",
			Unique: true,
			Attrs: []schema.Attr{
				&mysql.IndexType{
					T: "BTREE",
				},
			},
			Parts: []*schema.IndexPart{
				{
					SeqNo: 0,
					C:     table.Columns[0],
				},
			},
		},
	}
	return &schema.Schema{
		Name: "test",
		Tables: []*schema.Table{
			table,
		},
	}
}

func MockMySQLM2MTwoTypes() *schema.Schema {
	tableA := &schema.Table{
		Name: "groups",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	tableA.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     tableA.Columns[0],
			},
		},
	}
	tableB := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	tableB.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     tableB.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "group_users",
		Columns: []*schema.Column{
			{
				Name: "group_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.Indexes = []*schema.Index{
		{
			Name:   "group_users_user_id",
			Unique: false,
			Table:  joinTable,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     joinTable.Columns[1],
				},
			},
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     joinTable.Columns[0],
			},
			{
				SeqNo: 1,
				C:     joinTable.Columns[1],
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "group_users_group_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: tableA,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "group_users_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: tableB,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	return &schema.Schema{
		Name:   "m2m_two_types",
		Tables: []*schema.Table{tableA, tableB, joinTable},
	}
}

func MockMySQLM2MSameType() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "user_following",
		Columns: []*schema.Column{
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "follower_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.Indexes = []*schema.Index{
		{
			Name:   "user_following_follower_id",
			Unique: false,
			Table:  joinTable,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     joinTable.Columns[1],
				},
			},
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     joinTable.Columns[0],
			},
			{
				SeqNo: 1,
				C:     joinTable.Columns[1],
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "user_following_follower_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "user_following_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	return &schema.Schema{
		Name:   "m2m_same_type",
		Tables: []*schema.Table{table, joinTable},
	}
}

func MockMySQLM2MBidirectional() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "user_friends",
		Columns: []*schema.Column{
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "friend_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.Indexes = []*schema.Index{
		{
			Name:   "user_friends_friend_id",
			Unique: false,
			Table:  joinTable,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     joinTable.Columns[1],
				},
			},
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     joinTable.Columns[0],
			},
			{
				SeqNo: 1,
				C:     joinTable.Columns[1],
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "user_friends_friend_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "user_friends_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	return &schema.Schema{
		Name:   "m2m_bidirectional",
		Tables: []*schema.Table{table, joinTable},
	}
}

func MockMySQLO2OTwoTypes() *schema.Schema {
	parentTable := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
		},
		PrimaryKey: &schema.Index{
			Name:   "PRI",
			Unique: false,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 0,
					C: &schema.Column{
						Name: "id",
						Type: &schema.ColumnType{
							Type: &schema.IntegerType{
								T:        "bigint",
								Unsigned: false,
							},
							Raw:  "bigint",
							Null: false,
						},
						Attrs: []schema.Attr{
							&mysql.AutoIncrement{
								V: 0,
							},
						},
					},
				},
			},
		},
	}
	childTable := &schema.Table{
		Name: "cards",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "number",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
			{
				Name: "user_card",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: true,
				},
			},
		},
		PrimaryKey: &schema.Index{
			Name:   "PRI",
			Unique: false,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 0,
					C: &schema.Column{
						Name: "id",
						Type: &schema.ColumnType{
							Type: &schema.IntegerType{
								T:        "bigint",
								Unsigned: false,
							},
							Raw:  "bigint",
							Null: false,
						},
						Attrs: []schema.Attr{
							&mysql.AutoIncrement{
								V: 0,
							},
						},
					},
				},
			},
		},
	}
	indexes := []*schema.Index{
		{
			Name:   "user_card",
			Unique: true,
			Table:  childTable,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     childTable.Columns[2],
				},
			},
		},
	}
	fks := []*schema.ForeignKey{
		{
			Symbol: "cards_users_card",
			Table:  childTable,
			Columns: []*schema.Column{
				childTable.Columns[2],
			},
			RefTable: parentTable,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	childTable.ForeignKeys = fks
	childTable.Indexes = indexes
	return &schema.Schema{
		Name: "o2o_two_types",
		Tables: []*schema.Table{
			parentTable,
			childTable,
		},
	}
}

func MockMySQLO2OSameType() *schema.Schema {
	table := &schema.Table{
		Name: "nodes",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "value",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "node_next",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: true,
				},
			},
		},
	}
	fks := []*schema.ForeignKey{
		{
			Symbol: "nodes_nodes_next",
			Table:  table,
			Columns: []*schema.Column{
				table.Columns[2],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	indexes := []*schema.Index{
		{
			Name:   "node_next",
			Unique: true,
			Table:  table,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     table.Columns[2],
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[0],
			},
		},
	}
	table.Indexes = indexes
	table.ForeignKeys = fks
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "o2o_same_type",
		Tables: []*schema.Table{table},
	}
}

func MockMySQLO2OBidirectional() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
			{
				Name: "user_spouse",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: true,
				},
			},
		},
	}
	indexes := []*schema.Index{
		{
			Name:   "user_spouse",
			Unique: true,
			Table:  table,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     table.Columns[3],
				},
			},
		},
	}
	fks := []*schema.ForeignKey{
		{
			Symbol: "users_users_spouse",
			Table:  table,
			Columns: []*schema.Column{
				table.Columns[3],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[0],
			},
		},
	}
	table.Indexes = indexes
	table.ForeignKeys = fks
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "o2o_bidirectional",
		Tables: []*schema.Table{table},
	}
}

func MockMySQLO2MTwoTypes() *schema.Schema {
	parentTable := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     parentTable.Columns[0],
			},
		},
	}
	parentTable.PrimaryKey = primaryKey
	childTable := &schema.Table{
		Name: "pets",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
			{
				Name: "user_pets",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: true,
				},
			},
		},
	}
	indexes := []*schema.Index{
		{
			Name:   "pets_users_pets",
			Unique: false,
			Table:  childTable,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     childTable.Columns[2],
				},
			},
		},
	}
	fks := []*schema.ForeignKey{
		{
			RefTable: parentTable,
			Symbol:   "pets_users_pets",
			Table:    childTable,
			Columns: []*schema.Column{
				childTable.Columns[2],
			},
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	primaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     childTable.Columns[0],
			},
		},
	}
	childTable.ForeignKeys = fks
	childTable.PrimaryKey = primaryKey
	childTable.Indexes = indexes
	return &schema.Schema{
		Name:   "o2m_two_types",
		Tables: []*schema.Table{parentTable, childTable},
	}
}

func MockMySQLO2MSameType() *schema.Schema {
	table := &schema.Table{
		Name: "nodes",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "value",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "node_children",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: true,
				},
			},
		},
	}
	fks := []*schema.ForeignKey{
		{
			Symbol: "nodes_nodes_children",
			Table:  table,
			Columns: []*schema.Column{
				table.Columns[2],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	indexes := []*schema.Index{
		{
			Name:   "nodes_nodes_children",
			Unique: false,
			Table:  table,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     table.Columns[2],
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[0],
			},
		},
	}
	table.Indexes = indexes
	table.ForeignKeys = fks
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "o2m_same_type",
		Tables: []*schema.Table{table},
	}
}

func MockMySQLO2XOtherSideIgnored() *schema.Schema {
	parentTable := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&mysql.AutoIncrement{
						V: 0,
					},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Charset{V: "utf8mb4"},
					&schema.Collation{V: "utf8mb4_bin"},
				},
			},
		},
		PrimaryKey: &schema.Index{
			Name:   "PRI",
			Unique: false,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 0,
					C: &schema.Column{
						Name: "id",
						Type: &schema.ColumnType{
							Type: &schema.IntegerType{
								T:        "bigint",
								Unsigned: false,
							},
							Raw:  "bigint",
							Null: false,
						},
						Attrs: []schema.Attr{
							&mysql.AutoIncrement{
								V: 0,
							},
						},
					},
				},
			},
		},
	}
	childTable := &schema.Table{
		Name: "pets",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
			{
				Name: "user_pets",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: true,
				},
			},
		},
	}
	indexes := []*schema.Index{
		{
			Name:   "pets_users_pets",
			Unique: false,
			Table:  childTable,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     childTable.Columns[2],
				},
			},
		},
	}
	fks := []*schema.ForeignKey{
		{
			RefTable: parentTable,
			Symbol:   "pets_users_pets",
			Table:    childTable,
			Columns: []*schema.Column{
				childTable.Columns[2],
			},
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	primaryKey := &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     childTable.Columns[0],
			},
		},
	}
	childTable.ForeignKeys = fks
	childTable.PrimaryKey = primaryKey
	childTable.Indexes = indexes
	return &schema.Schema{
		Name:   "o2m_two_types",
		Tables: []*schema.Table{childTable},
	}
}

func MockMySQLM2MJoinTableOnly() *schema.Schema {
	tableA := &schema.Table{
		Name: "groups",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	tableA.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     tableA.Columns[0],
			},
		},
	}
	tableB := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
					Raw:  "varchar(255)",
					Null: false,
				},
			},
		},
	}
	tableB.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     tableB.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "group_users",
		Columns: []*schema.Column{
			{
				Name: "group_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.Indexes = []*schema.Index{
		{
			Name:   "group_users_user_id",
			Unique: false,
			Table:  joinTable,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     joinTable.Columns[1],
				},
			},
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     joinTable.Columns[0],
			},
			{
				SeqNo: 1,
				C:     joinTable.Columns[1],
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "group_users_group_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: tableA,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "group_users_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: tableB,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	return &schema.Schema{
		Name:   "m2m_two_types",
		Tables: []*schema.Table{joinTable},
	}
}

func MockPostgresSingleTableFields() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "smallint",
						Unsigned: false,
					},
					Raw:  "smallint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  table,
		Attrs: []schema.Attr{
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "public",
		Tables: []*schema.Table{table},
	}
}

func MockPostgresTableFieldsWithAttributes() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "some id"},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "smallint",
						Unsigned: false,
					},
					Raw:  "smallint",
					Null: true,
				},
				Default: &schema.RawExpr{X: "1"},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "first name"},
				},
			},
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: true,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "family name"},
				},
			},
		},
	}
	primaryKey := &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  table,
		Attrs: []schema.Attr{
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	table.PrimaryKey = primaryKey
	return &schema.Schema{
		Name:   "public",
		Tables: []*schema.Table{table},
	}
}

func MockPostgresTableFieldsWithUniqueIndexes() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "some id"},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "smallint",
						Unsigned: false,
					},
					Raw:  "smallint",
					Null: false,
				},
				Default: &schema.RawExpr{X: "1"},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "first name"},
				},
			},
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: true,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "family name"},
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Table:  table,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     table.Columns[0],
			},
		},
	}
	table.Indexes = []*schema.Index{
		{
			Name:   "users_age_uindex",
			Unique: true,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: table.Columns[1],
				},
			},
		},
	}
	return &schema.Schema{
		Name:   "test",
		Tables: []*schema.Table{table},
	}
}

func MockPostgresMultiTableFields() *schema.Schema {
	tableA := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "smallint",
						Unsigned: false,
					},
					Raw:  "smallint",
					Null: false,
				},
				Default: &schema.RawExpr{X: "1"},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: true,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "not so boring"},
				},
			},
		},
	}
	tableA.Indexes = []*schema.Index{
		{
			Name:   "users_age_uindex",
			Unique: true,
			Table:  tableA,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: tableA.Columns[1]},
			},
		},
	}
	tableA.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  tableA,
		Attrs: []schema.Attr{
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: tableA.Columns[0],
			},
		},
	}
	tableB := &schema.Table{
		Name: "pets",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "pet id"},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "smallint",
						Unsigned: false,
					},
					Raw:  "smallint",
					Null: true,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	tableB.PrimaryKey = &schema.Index{
		Name:   "PRI",
		Unique: false,
		Table:  tableB,
		Parts: []*schema.IndexPart{
			{
				SeqNo: 0,
				C:     tableB.Columns[0],
			},
		},
	}
	return &schema.Schema{
		Name: "test",
		Tables: []*schema.Table{
			tableA,
			tableB,
		},
	}
}

func MockPostgresNonDefaultPrimaryKey() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
			{
				Name: "last_name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: true,
				},
				Attrs: []schema.Attr{
					&schema.Comment{Text: "not so boring"},
				},
			},
		},
	}
	table.Indexes = []*schema.Index{
		{
			Name:   "users_last_name_uindex",
			Unique: true,
			Table:  table,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: table.Columns[1],
				},
			},
		},
		{
			Name:   "users_name_index",
			Unique: false,
			Table:  table,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: table.Columns[0],
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  table,
		Attrs: []schema.Attr{
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	return &schema.Schema{
		Name:   "test",
		Tables: []*schema.Table{table},
	}
}

func MockPostgresNonDefaultPrimaryKeyWithIndexes() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "my_id",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	table.Indexes = []*schema.Index{
		{
			Name:   "users_my_id_uindex",
			Unique: true,
			Table:  table,
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: table.Columns[0],
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  table,
		Attrs: []schema.Attr{
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	return &schema.Schema{
		Name:   "test",
		Tables: []*schema.Table{table},
	}
}

func MockPostgresM2MTwoTypes() *schema.Schema {
	tableA := &schema.Table{
		Name: "groups",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	tableA.PrimaryKey = &schema.Index{
		Name:   "groups_pkey",
		Unique: true,
		Table:  tableA,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: tableA.Columns[0],
			},
		},
	}
	tableB := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	tableB.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  tableB,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: tableB.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "group_users",
		Columns: []*schema.Column{
			{
				Name: "group_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "group_users_group_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: tableA,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "group_users_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: tableB,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "group_users_pkey",
		Unique: true,
		Table:  joinTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[0],
			},
			{
				SeqNo: 2,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[1],
			},
		},
	}
	return &schema.Schema{
		Name:   "m2m_two_types",
		Tables: []*schema.Table{tableA, tableB, joinTable},
	}
}

func MockPostgresM2MSameType() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  (*schema.Table)(nil),
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "user_following",
		Columns: []*schema.Column{
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "follower_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "user_following_pkey",
		Unique: true,
		Table:  joinTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[0],
			},
			{
				SeqNo: 2,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[1],
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "user_following_follower_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "user_following_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	return &schema.Schema{
		Name:   "m2m_same_type",
		Tables: []*schema.Table{joinTable, table},
	}
}

func MockPostgresM2MBidirectional() *schema.Schema {
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  table,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "user_friends",
		Columns: []*schema.Column{
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "friend_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.Indexes = []*schema.Index{
		{
			Name:   "user_friends_friend_id",
			Unique: false,
			Table:  (*schema.Table)(nil),
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					C:     joinTable.Columns[1],
				},
			},
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "user_friends_pkey",
		Unique: true,
		Table:  joinTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[0],
			},
			{
				SeqNo: 2,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[1],
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "user_friends_friend_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "user_friends_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	return &schema.Schema{
		Name:   "m2m_bidirectional",
		Tables: []*schema.Table{table, joinTable},
	}
}

func MockPostgresO2OTwoTypes() *schema.Schema {
	parentTable := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	parentTable.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  (*schema.Table)(nil),
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: parentTable.Columns[0],
			},
		},
	}
	childTable := &schema.Table{
		Name: "cards",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "expired",
				Type: &schema.ColumnType{
					Type: &schema.TimeType{T: "timestamp with time zone"},
					Raw:  "timestamp with time zone",
					Null: false,
				},
			},
			{
				Name: "number",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
			{
				Name: "user_card",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: true,
				},
			},
		},
	}
	childTable.PrimaryKey = &schema.Index{
		Name:   "cards_pkey",
		Unique: true,
		Table:  childTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: childTable.Columns[0],
			},
		},
	}
	childTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol:   "cards_users_card",
			Table:    childTable,
			RefTable: parentTable,
			Columns:  []*schema.Column{childTable.Columns[3]},
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	childTable.Indexes = []*schema.Index{
		{
			Name:   "cards_user_card_key",
			Unique: true,
			Table:  (*schema.Table)(nil),
			Attrs: []schema.Attr{
				&postgres.IndexType{
					T: "btree",
				},
				&postgres.ConType{
					T: "u",
				},
			},
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: childTable.Columns[3],
				},
			},
		},
	}
	return &schema.Schema{
		Name: "o2o_two_types",
		Tables: []*schema.Table{
			parentTable,
			childTable,
		},
	}
}

func MockPostgresO2OSameType() *schema.Schema {
	table := &schema.Table{
		Name: "nodes",
	}
	table.Columns = []*schema.Column{
		{
			Name: "id",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
			Attrs: []schema.Attr{
				&postgres.Identity{
					Generation: "BY DEFAULT",
				},
			},
		},
		{
			Name: "value",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
		},
		{
			Name: "node_next",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: true,
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "nodes_pkey",
		Unique: true,
		Table:  (*schema.Table)(nil),
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	table.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "nodes_nodes_next",
			Table:  table,
			Columns: []*schema.Column{
				table.Columns[2],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	table.Indexes = []*schema.Index{
		{
			Name:   "nodes_node_next_key",
			Unique: true,
			Table:  (*schema.Table)(nil),
			Attrs: []schema.Attr{
				&postgres.IndexType{
					T: "btree",
				},
				&postgres.ConType{
					T: "u",
				},
			},
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: table.Columns[2],
				},
			},
		},
	}
	return &schema.Schema{
		Name: "o2o_two_types",
		Tables: []*schema.Table{
			table,
		},
	}
}

func MockPostgresO2OBidirectional() *schema.Schema {
	table := &schema.Table{
		Name: "user",
	}
	table.Columns = []*schema.Column{
		{
			Name: "id",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
			Attrs: []schema.Attr{
				&postgres.Identity{
					Generation: "BY DEFAULT",
				},
			},
		},
		{
			Name: "age",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
		},
		{
			Name: "name",
			Type: &schema.ColumnType{
				Type: &schema.StringType{T: "character varying", Size: 0},
				Raw:  "character varying",
				Null: false,
			},
		},
		{
			Name: "user_spouse",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: true,
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  table,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	table.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "users_users_spouse",
			Table:  table,
			Columns: []*schema.Column{
				table.Columns[3],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	table.Indexes = []*schema.Index{
		{
			Name:   "users_user_spouse_key",
			Unique: true,
			Table:  table,
			Attrs: []schema.Attr{
				&postgres.IndexType{
					T: "btree",
				},
				&postgres.ConType{
					T: "u",
				},
			},
			Parts: []*schema.IndexPart{
				{
					SeqNo: 1,
					Attrs: []schema.Attr{
						&postgres.IndexColumnProperty{
							NullsFirst: false,
							NullsLast:  true,
						},
					},
					C: table.Columns[3],
				},
			},
		},
	}
	return &schema.Schema{
		Name: "o2o_bidirectional",
		Tables: []*schema.Table{
			table,
		},
	}
}

func MockPostgresO2MTwoTypes() *schema.Schema {
	parentTable := &schema.Table{
		Name: "users",
	}
	parentTable.Columns = []*schema.Column{
		{
			Name: "id",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
			Attrs: []schema.Attr{
				&postgres.Identity{
					Generation: "BY DEFAULT",
				},
			},
		},
		{
			Name: "age",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
		},
		{
			Name: "name",
			Type: &schema.ColumnType{
				Type: &schema.StringType{T: "character varying", Size: 0},
				Raw:  "character varying",
				Null: false,
			},
		},
	}
	parentTable.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  parentTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: parentTable.Columns[0],
			},
		},
	}
	childTable := &schema.Table{
		Name: "pets",
	}
	childTable.Columns = []*schema.Column{
		{
			Name: "id",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
			Attrs: []schema.Attr{
				&postgres.Identity{
					Generation: "BY DEFAULT",
				},
			},
		},
		{
			Name: "name",
			Type: &schema.ColumnType{
				Type: &schema.StringType{T: "character varying", Size: 0},
				Raw:  "character varying",
				Null: false,
			},
		},
		{
			Name: "user_pets",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: true,
			},
		},
	}
	childTable.PrimaryKey = &schema.Index{
		Name:   "pets_pkey",
		Unique: true,
		Table:  childTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: childTable.Columns[0],
			},
		},
	}
	childTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "pets_users_pets",
			Table:  childTable,
			Columns: []*schema.Column{
				childTable.Columns[2],
			},
			RefTable: parentTable,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	return &schema.Schema{
		Name:   "o2m_two_types",
		Tables: []*schema.Table{parentTable, childTable},
	}
}

func MockPostgresO2MSameType() *schema.Schema {
	table := &schema.Table{
		Name: "nodes",
	}
	table.Columns = []*schema.Column{
		{
			Name: "id",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
			Attrs: []schema.Attr{
				&postgres.Identity{
					Generation: "BY DEFAULT",
				},
			},
		},
		{
			Name: "value",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
		},
		{
			Name: "node_children",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: true,
			},
		},
	}
	table.PrimaryKey = &schema.Index{
		Name:   "nodes_pkey",
		Unique: true,
		Table:  (*schema.Table)(nil),
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: table.Columns[0],
			},
		},
	}
	table.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "nodes_nodes_children",
			Table:  table,
			Columns: []*schema.Column{
				table.Columns[2],
			},
			RefTable: table,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	return &schema.Schema{
		Name:   "o2m_same_type",
		Tables: []*schema.Table{table},
	}
}

func MockPostgresO2XOtherSideIgnored() *schema.Schema {
	parentTable := &schema.Table{
		Name: "users",
	}
	parentTable.Columns = []*schema.Column{
		{
			Name: "id",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
			Attrs: []schema.Attr{
				&postgres.Identity{
					Generation: "BY DEFAULT",
				},
			},
		},
		{
			Name: "age",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
		},
		{
			Name: "name",
			Type: &schema.ColumnType{
				Type: &schema.StringType{T: "character varying", Size: 0},
				Raw:  "character varying",
				Null: false,
			},
		},
	}
	parentTable.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  parentTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: parentTable.Columns[0],
			},
		},
	}
	childTable := &schema.Table{
		Name: "pets",
	}
	childTable.Columns = []*schema.Column{
		{
			Name: "id",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: false,
			},
			Attrs: []schema.Attr{
				&postgres.Identity{
					Generation: "BY DEFAULT",
				},
			},
		},
		{
			Name: "name",
			Type: &schema.ColumnType{
				Type: &schema.StringType{T: "character varying", Size: 0},
				Raw:  "character varying",
				Null: false,
			},
		},
		{
			Name: "user_pets",
			Type: &schema.ColumnType{
				Type: &schema.IntegerType{
					T:        "bigint",
					Unsigned: false,
				},
				Raw:  "bigint",
				Null: true,
			},
		},
	}
	childTable.PrimaryKey = &schema.Index{
		Name:   "pets_pkey",
		Unique: true,
		Table:  childTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: childTable.Columns[0],
			},
		},
	}
	childTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "pets_users_pets",
			Table:  childTable,
			Columns: []*schema.Column{
				childTable.Columns[2],
			},
			RefTable: parentTable,
			OnUpdate: "NO ACTION",
			OnDelete: "SET NULL",
		},
	}
	return &schema.Schema{
		Name:   "o2m_two_types",
		Tables: []*schema.Table{childTable},
	}
}

func MockPostgresM2MJoinTableOnly() *schema.Schema {
	tableA := &schema.Table{
		Name: "groups",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	tableA.PrimaryKey = &schema.Index{
		Name:   "groups_pkey",
		Unique: true,
		Table:  tableA,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: tableA.Columns[0],
			},
		},
	}
	tableB := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
				Attrs: []schema.Attr{
					&postgres.Identity{
						Generation: "BY DEFAULT",
					},
				},
			},
			{
				Name: "age",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "character varying", Size: 0},
					Raw:  "character varying",
					Null: false,
				},
			},
		},
	}
	tableB.PrimaryKey = &schema.Index{
		Name:   "users_pkey",
		Unique: true,
		Table:  tableB,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: tableB.Columns[0],
			},
		},
	}
	joinTable := &schema.Table{
		Name: "group_users",
		Columns: []*schema.Column{
			{
				Name: "group_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
			{
				Name: "user_id",
				Type: &schema.ColumnType{
					Type: &schema.IntegerType{
						T:        "bigint",
						Unsigned: false,
					},
					Raw:  "bigint",
					Null: false,
				},
			},
		},
	}
	joinTable.ForeignKeys = []*schema.ForeignKey{
		{
			Symbol: "group_users_group_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[0],
			},
			RefTable: tableA,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
		{
			Symbol: "group_users_user_id",
			Table:  joinTable,
			Columns: []*schema.Column{
				joinTable.Columns[1],
			},
			RefTable: tableB,
			OnUpdate: "NO ACTION",
			OnDelete: "CASCADE",
		},
	}
	joinTable.PrimaryKey = &schema.Index{
		Name:   "group_users_pkey",
		Unique: true,
		Table:  joinTable,
		Attrs: []schema.Attr{
			&postgres.IndexType{
				T: "btree",
			},
			&postgres.ConType{
				T: "p",
			},
		},
		Parts: []*schema.IndexPart{
			{
				SeqNo: 1,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[0],
			},
			{
				SeqNo: 2,
				Attrs: []schema.Attr{
					&postgres.IndexColumnProperty{
						NullsFirst: false,
						NullsLast:  true,
					},
				},
				C: joinTable.Columns[1],
			},
		},
	}
	return &schema.Schema{
		Name:   "m2m_two_types",
		Tables: []*schema.Table{joinTable},
	}
}

// Inspector is an autogenerated mock type for the Inspector type
type inspectorMock struct {
	mock.Mock
}

func (_m *inspectorMock) InspectTable(_ context.Context, _ string, _ *schema.InspectTableOptions) (*schema.Table, error) {
	return nil, nil
}

func (_m *inspectorMock) InspectRealm(_ context.Context, _ *schema.InspectRealmOption) (*schema.Realm, error) {
	return nil, nil
}

// InspectSchema provides a mock function with given fields: ctx, name, opts
func (_m *inspectorMock) InspectSchema(ctx context.Context, name string, opts *schema.InspectOptions) (*schema.Schema, error) {
	ret := _m.Called(ctx, name, opts)
	var r0 *schema.Schema
	if rf, ok := ret.Get(0).(func(context.Context, string, *schema.InspectOptions) *schema.Schema); ok {
		r0 = rf(ctx, name, opts)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*schema.Schema)
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *schema.InspectOptions) error); ok {
		r1 = rf(ctx, name, opts)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func createTempDir(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "entimport-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		err = os.RemoveAll(tmpDir)
		require.NoError(t, err)
	})
	return tmpDir
}

func readDir(t *testing.T, path string) map[string]string {
	files := make(map[string]string)
	err := filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}
		buf, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.Base(path)] = string(buf)
		return nil
	})
	require.NoError(t, err)
	return files
}

func lookupMethod(file *ast.File, typeName string, methodName string) (m *ast.FuncDecl) {
	ast.Inspect(file, func(node ast.Node) bool {
		if decl, ok := node.(*ast.FuncDecl); ok {
			if decl.Name.Name != methodName || decl.Recv == nil || len(decl.Recv.List) != 1 {
				return true
			}
			if id, ok := decl.Recv.List[0].Type.(*ast.Ident); ok && id.Name == typeName {
				m = decl
				return false
			}
		}
		return true
	})
	return m
}

func mockMux(ctx context.Context, dlct string, data *schema.Schema, schemaName string) *mux.Mux {
	im := &inspectorMock{}
	im.On("InspectSchema", ctx, schemaName, &schema.InspectOptions{}).Return(data, nil)
	m := mux.New()
	m.RegisterProvider(func(s string) (*mux.ImportDriver, error) {
		return &mux.ImportDriver{
			Inspector:  im,
			Dialect:    dlct,
			SchemaName: schemaName,
		}, nil
	}, dlct)
	return m
}
