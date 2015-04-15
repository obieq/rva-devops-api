package rethinkdb

import (
	"log"
	"strings"

	r "github.com/dancannon/gorethink"
)

type Migrator interface {
	CreateDb(dbName string) error
	DropDb(dbName string) error
	CreateTable(tableName string) error
	DropTable(tableName string) error
	AddIndex(tableName string, fields []string, opts map[string]interface{}) error
}

type RethinkDbMigration struct {
	Migrator
}

func (*RethinkDbMigration) CreateDb(dbName string) error {
	if _, err := r.DbCreate(dbName).Run(Session()); err != nil {
		//log.Fatalln(err.Error())
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*RethinkDbMigration) DropDb(dbName string) error {
	if _, err := r.DbDrop(dbName).Run(Session()); err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*RethinkDbMigration) CreateTable(tableName string) error {
	if _, err := r.Db(DbName()).TableCreate(tableName).RunWrite(Session()); err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*RethinkDbMigration) DropTable(tableName string) error {
	if _, err := r.Db(DbName()).TableDrop(tableName).RunWrite(Session()); err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (*RethinkDbMigration) AddIndex(tableName string, fields []string, opts map[string]interface{}) error {
	if len(fields) == 1 {
		if _, err := r.Db(DbName()).Table(tableName).IndexCreate(fields[0], r.IndexCreateOpts{Multi: true}).Run(Session()); err != nil {
			log.Println(err.Error())
			return err
		}
	} else {
		indexName := strings.Join(fields, "_")
		if _, err := r.Db(DbName()).Table(tableName).IndexCreateFunc(indexName, func(row r.Term) interface{} {
			fieldSlice := []r.Term{}
			for _, element := range fields {
				fieldSlice = append(fieldSlice, row.Field(element))
			}

			return []interface{}{fieldSlice}
		}).RunWrite(Session()); err != nil {
			log.Println(err.Error())
			return err
		}
	}

	return nil
}
