package rethinkdb

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RethinkDb Migration", func() {

	var (
		Migration          *RethinkDbMigration = &RethinkDbMigration{}
		migrationDbName    string              = "migration_test_db"
		migrationTableName string              = "migration_test_table"
	)

	Context("Setup", func() {
		It("should create a database", func() {
			err := Migration.CreateDb(migrationDbName)
			Ω(err).NotTo(HaveOccurred())
		})

		It("should create a table", func() {
			err := Migration.CreateTable(migrationTableName)
			Ω(err).NotTo(HaveOccurred())
		})

		It("should create a single column index", func() {
			err := Migration.AddIndex(migrationTableName, []string{"age"}, nil)
			Ω(err).NotTo(HaveOccurred())
		})

		It("should create a multiple column index", func() {
			err := Migration.AddIndex(migrationTableName, []string{"first_name", "last_name"}, nil)
			Ω(err).NotTo(HaveOccurred())
		})
	})

	Context("Error Handling", func() {
		It("should return an error when creating a database with an invalid name", func() {
			err := Migration.CreateDb("")
			Ω(err).To(HaveOccurred())
		})

		It("should raise an error when creating a database using an existing name", func() {
			err := Migration.CreateDb(migrationDbName)
			Ω(err).To(HaveOccurred())
		})

		It("should raise an error when dropping a non-existent database", func() {
			err := Migration.DropDb("lorem_ipsum")
			Ω(err).To(HaveOccurred())
		})

		It("should return an error when creating a table with an invalid name", func() {
			err := Migration.CreateTable("")
			Ω(err).To(HaveOccurred())
		})

		It("should raise an error when creating a table using an existing name", func() {
			err := Migration.CreateTable(migrationTableName)
			Ω(err).To(HaveOccurred())
		})

		It("should raise an error when dropping a non-existent table", func() {
			err := Migration.DropTable("lorem_ipsum")
			Ω(err).To(HaveOccurred())
		})

		It("should raise an error when creating an single field index using an existing name", func() {
			err := Migration.AddIndex(migrationTableName, []string{"age"}, nil)
			Ω(err).To(HaveOccurred())
		})

		It("should raise an error when creating a multiple column index using an existing name", func() {
			err := Migration.AddIndex(migrationTableName, []string{"first_name", "last_name"}, nil)
			Ω(err).To(HaveOccurred())
		})
	})

	Context("Cleanup", func() {
		It("should drop a table", func() {
			err := Migration.DropTable(migrationTableName)
			Ω(err).NotTo(HaveOccurred())
		})

		It("should drop a database", func() {
			err := Migration.DropDb(migrationDbName)
			Ω(err).NotTo(HaveOccurred())
		})
	})
})
