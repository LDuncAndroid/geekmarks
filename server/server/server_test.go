package server

import (
	"database/sql"
	"flag"
	"os"
	"strings"
	"testing"

	"dmitryfrank.com/geekmarks/server/storage"
	"github.com/golang/glog"
	"github.com/juju/errors"
)

func TestMain(m *testing.M) {
	flag.Parse()

	err := Initialize(false /*don't apply migrations*/)
	if err != nil {
		glog.Fatalf("%s\n", errors.ErrorStack(err))
	}

	os.Exit(m.Run())
}

func TestServer(t *testing.T) {
	err := dbPrepare(t)
	if err != nil {
		t.Errorf("%s", err)
	}

	err = dbCleanup(t)
	if err != nil {
		t.Errorf("%s", err)
	}
}

func getAllTables(t *testing.T) ([]string, error) {
	var tables []string
	err := storage.Tx(func(tx *sql.Tx) error {
		rows, err := tx.Query(`
			SELECT table_name
				FROM information_schema.tables
				WHERE table_schema='public'
				AND table_type='BASE TABLE'
		`)
		if err != nil {
			return errors.Trace(err)
		}
		defer rows.Close()
		for rows.Next() {
			var tableName string
			err := rows.Scan(&tableName)
			if err != nil {
				return errors.Trace(err)
			}

			tables = append(tables, tableName)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Annotatef(err, "dropping all tables")
	}

	return tables, nil
}

func dbPrepare(t *testing.T) error {
	// Drop all existing tables
	tables, err := getAllTables(t)
	if err != nil {
		return errors.Annotatef(err, "getting all table names")
	}

	if len(tables) > 0 {
		err = storage.Tx(func(tx *sql.Tx) error {
			q := "DROP TABLE " + strings.Join(tables, ", ")
			t.Log(q)
			_, err = tx.Exec(q)
			if err != nil {
				return errors.Annotatef(err, "dropping all tables")
			}

			return nil
		})
		if err != nil {
			return errors.Trace(err)
		}
	}

	// Init schema (apply all migrations)
	err = storage.ApplyMigrations()
	if err != nil {
		return errors.Annotatef(err, "applying migrations")
	}

	return nil
}

func dbCleanup(t *testing.T) error {
	// TODO: migrate down and check that no tables are present
	return nil
}
