package entimport_test

import (
	"context"
	"os"
	"testing"

	"ariga.io/atlas/sql/schema"
	"ariga.io/entimport/internal/entimport"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/suite"
)

const testSchema = "test"

type MySQLImportTestSuite struct {
	suite.Suite
	importer   *entimport.MySQL
	schemaPath string
}

func TestMySQLSuite(t *testing.T) {
	suite.Run(t, new(MySQLImportTestSuite))
}

func (s *MySQLImportTestSuite) SetupTest() {
	tempDir := createTempDir(s.T())
	i := &entimport.ImportOptions{}
	entimport.WithDSN("root:pass@tcp(localhost:3308)/test?parseTime=True")(i)
	entimport.WithSchemaPath(tempDir)(i)
	s.importer = &entimport.MySQL{Options: i}
	s.schemaPath = tempDir
}

func (s *MySQLImportTestSuite) TestSingleTableFields() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLSingleTableFields(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	expectedSchema, err := os.ReadFile("../testdata/fields/singletable/user.go")
	s.NoError(err)
	userSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	s.EqualValues(expectedSchema, userSchema)
}

func (s *MySQLImportTestSuite) TestTableFieldsWithAttributes() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLTableFieldsWithAttributes(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	expectedSchema, err := os.ReadFile("../testdata/fields/tablefields/user.go")
	s.NoError(err)
	userSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	s.EqualValues(expectedSchema, userSchema)
}

func (s *MySQLImportTestSuite) TestTableFieldsWithUniqueIndexes() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLTableFieldsWithUniqueIndexes(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	expectedSchema, err := os.ReadFile("../testdata/fields/uniqueindex/user.go")
	s.NoError(err)
	userSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), userSchema)
}

func (s *MySQLImportTestSuite) TestMultiTableFields() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLMultiTableFields(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 2)
	expectedSchema, err := os.ReadFile("../testdata/fields/multitable/user.go")
	s.NoError(err)
	userSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), userSchema)
	expectedSchema, err = os.ReadFile("../testdata/fields/multitable/pet.go")
	s.NoError(err)
	petSchema, ok := schemaFiles["pet.go"]
	s.True(ok)
	s.EqualValues(expectedSchema, petSchema)
}

func (s *MySQLImportTestSuite) TestNonDefaultPrimaryKey() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLNonDefaultPrimaryKey(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 1)
	userSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	expectedSchema, err := os.ReadFile("../testdata/fields/primarykey/user.go")
	s.NoError(err)
	s.EqualValues(string(expectedSchema), userSchema)
}

func (s *MySQLImportTestSuite) TestRelationM2MTwoTypes() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLM2MTwoTypes(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 2)
	expectedSchema, err := os.ReadFile("../testdata/relations/m2m2types/group.go")
	s.NoError(err)
	parentSchema, ok := schemaFiles["group.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), parentSchema)
	expectedSchema, err = os.ReadFile("../testdata/relations/m2m2types/user.go")
	s.NoError(err)
	childSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), childSchema)
}

func (s *MySQLImportTestSuite) TestRelationM2MSameType() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLM2MSameType(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 1)
	actualSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	expectedSchema, err := os.ReadFile("../testdata/relations/m2mrecur/user.go")
	s.NoError(err)
	s.EqualValues(string(expectedSchema), actualSchema)
}

func (s *MySQLImportTestSuite) TestRelationM2MBidirectional() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLM2MBidirectional(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 1)
	actualSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	expectedSchema, err := os.ReadFile("../testdata/relations/m2mbidi/user.go")
	s.NoError(err)
	s.EqualValues(string(expectedSchema), actualSchema)
}

func (s *MySQLImportTestSuite) TestRelationO2OTwoTypes() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLO2OTwoTypes(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 2)
	expectedSchema, err := os.ReadFile("../testdata/relations/o2o2types/user.go")
	s.NoError(err)
	userSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), userSchema)
	expectedSchema, err = os.ReadFile("../testdata/relations/o2o2types/card.go")
	s.NoError(err)
	cardSchema, ok := schemaFiles["card.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), cardSchema)
}

func (s *MySQLImportTestSuite) TestRelationO2OSameType() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLO2OSameType(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 1)
	actualSchema, ok := schemaFiles["node.go"]
	s.True(ok)
	expectedSchema, err := os.ReadFile("../testdata/relations/o2orecur/node.go")
	s.NoError(err)
	s.EqualValues(string(expectedSchema), actualSchema)
}

func (s *MySQLImportTestSuite) TestRelationO2OBidirectional() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLO2OBidirectional(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 1)
	actualSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	expectedSchema, err := os.ReadFile("../testdata/relations/o2obidi/user.go")
	s.NoError(err)
	s.EqualValues(string(expectedSchema), actualSchema)
}

func (s *MySQLImportTestSuite) TestRelationO2MTwoTypes() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLO2MTwoTypes(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 2)
	expectedSchema, err := os.ReadFile("../testdata/relations/o2m2types/user.go")
	s.NoError(err)
	parentSchema, ok := schemaFiles["user.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), parentSchema)
	expectedSchema, err = os.ReadFile("../testdata/relations/o2m2types/pet.go")
	s.NoError(err)
	childSchema, ok := schemaFiles["pet.go"]
	s.True(ok)
	s.EqualValues(string(expectedSchema), childSchema)
}

func (s *MySQLImportTestSuite) TestRelationO2MSameType() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLO2MSameType(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 1)
	actualSchema, ok := schemaFiles["node.go"]
	s.True(ok)
	expectedSchema, err := os.ReadFile("../testdata/relations/o2mrecur/node.go")
	s.NoError(err)
	s.EqualValues(string(expectedSchema), actualSchema)
}

// Case the `-tables` flag didn't provide all sides of the relation.
func (s *MySQLImportTestSuite) TestRelationO2XOtherSideIgnored() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLO2XOtherSideIgnored(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.NoError(err)
	err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(s.schemaPath))
	s.NoError(err)
	schemaFiles := readDir(s.T(), s.schemaPath)
	s.Len(schemaFiles, 1)
	actualSchema, ok := schemaFiles["pet.go"]
	s.True(ok)
	expectedSchema, err := os.ReadFile("../testdata/relations/o2xignore/pet.go")
	s.NoError(err)
	s.EqualValues(string(expectedSchema), actualSchema)
}

func (s *MySQLImportTestSuite) TestRelationM2MJoinTableOnly() {
	ctx := context.Background()
	iMock := &Inspector{}
	iMock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(MockMySQLM2MJoinTableOnly(), nil)
	s.importer.Inspector = iMock
	mutations, err := s.importer.SchemaMutations(ctx)
	s.Empty(mutations)
	s.EqualError(err, "entimport: join tables must be inspected with ref tables - append `tables` flag")
}
