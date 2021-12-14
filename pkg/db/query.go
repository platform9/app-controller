package db

import (
	"fmt"
	// MySQL drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/platform9/fast-path/pkg/options"

	"database/sql"
	"sync"
)

// Querier facilitates querying appbert objects from DB
type Querier struct {
	handle *sql.DB
}

var dbInstance *Querier

//var pkgDBInstance *Querier
var onceDB sync.Once


// Get returns handle to DB
//go:generate go-bindata -pkg db -o migrations_generated.go schema/
func Get() *Querier {
	onceDB.Do(func() {
		var db *sql.DB
		var err error
		switch dbType := options.GetDBType(); dbType {
		case "sqlite3":
			db, err = sql.Open(dbType, options.GetDBSrc())
		case "mysql":
			db, err = sql.Open(dbType, options.GetDBCreds())
		default:
			panic(fmt.Sprintf("DB type %s is not supported", dbType))
		}

		if err != nil {
			panic(err)
		}

		if err = db.Ping(); err != nil {
			panic(err)
		}

		dbInstance = &Querier{
			handle: db,
		}
	})

	return dbInstance
}

// GetHandle returns DB handle
func (db *Querier) GetHandle() *sql.DB {
	return db.handle
}

func (db *Querier) DropData() {
	tx, err := db.handle.Begin()
	if err != nil {
		panic(err)
	}

	for _, t := range []string{"repos", "subs"} {
		stmt, err := tx.Prepare(fmt.Sprintf("delete from %s", t))
		if err != nil {
			panic(err)
		}
		if _, err = stmt.Exec(); err != nil {
			panic(err)
		}
	}

	tx.Commit()
}

// Migrate initializes and upgrades database
func (db *Querier) Migrate() error {
	return RunMigrations(db.handle)
}
