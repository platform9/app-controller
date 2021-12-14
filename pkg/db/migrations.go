package db

import (
	"log"
	"sort"
	"strings"

	"database/sql"
)

// RunMigrations runs DB migrations at startup
func RunMigrations(db *sql.DB) error {
	if err := verifyMigrationsTable(db); err != nil {
		return err
	}

	count, err := countMigrations(db)
	if err != nil {
		return err
	}

	names := AssetNames()
	sort.Strings(names)

	log.Printf("Files: %s", names)
	for i, file := range names {
		// skip running ones we've clearly already ran
		if count > 0 {
			count--
			continue
		}

		migration, err := Asset(file)
		if err != nil {
			return err
		}

		log.Printf("Running %s", file)
		err = runMigration(i, migration, db)
		if err != nil {
			return err
		}

		log.Printf("Migrations ran successfully %s", file)
		cleanName := strings.TrimPrefix(file, "schema/")
		err = recordMigration(cleanName, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func verifyMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY /*!40101 AUTO_INCREMENT*/,
		created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
		name VARCHAR(128) NOT NULL UNIQUE
	  );`)
	return err
}

func countMigrations(db *sql.DB) (int, error) {
	row := db.QueryRow(`SELECT count(*) as count FROM migrations;`)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func runMigration(num int, buf []byte, db *sql.DB) error {
	// TODO: split queries, in single file.
	_, err := db.Exec(string(buf))
	return err
}

func recordMigration(name string, db *sql.DB) error {
	query := `INSERT INTO migrations (name) VALUES (?);`
	_, err := db.Exec(query, name)
	return err
}
