package rethinkdb

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRethinkDb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RethinkDb Suite")
}

var Migration *RethinkDbMigration = &RethinkDbMigration{}
var migrationDbName string = "migration_test_db"

var _ = BeforeSuite(func() {
	dbName := "goar_test"
	SetDbName(dbName)

	// drop databases from prior test(s)
	err := Migration.DropDb(DbName())
	Migration.DropDb(migrationDbName)

	// prep for current test(s)
	err = Migration.CreateDb(DbName())
	Expect(err).NotTo(HaveOccurred())

	err = Migration.CreateTable("rethink_db_automobiles")
	Expect(err).NotTo(HaveOccurred())

	err = Migration.CreateTable("callback_error_models")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Session().Close()
})
