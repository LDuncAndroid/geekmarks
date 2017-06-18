// Copyright 2017 Dmitry Frank <mail@dmitryfrank.com>
// Licensed under the BSD, see LICENSE file for details.

package dfmigrate

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/juju/errors"
	_ "github.com/lib/pq"
)

const (
	paramCurMigrationID = "cur_migration_id"
)

func tx(db *sql.DB, fn func(*sql.Tx) error) error {
	var tx *sql.Tx
	var err error

	// Because of the way PostgreSQL container is designed, when it runs for the
	// first time, it's not immediately ready to accept connections: it needs
	// several seconds to bootstrap the database first. So here we use a timeout
	// hack: we keep retrying to connect for 10 seconds.
	timeoutChan := time.After(10 * time.Second)
	for {
		tx, err = db.Begin()
		if err != nil {
			err2 := errors.Annotate(err, "begin transaction")
			pqerr, ok := err.(*net.OpError)
			if ok {
				if pqerr.Err.Error() == "read: connection reset by peer" ||
					pqerr.Err.Error() == "getsockopt: connection refused" {
					fmt.Printf("Waiting more before connecting...\n")
					select {
					case <-timeoutChan:
						return errors.Annotate(err2, "time is out")
					case <-time.After(1 * time.Second):
						continue
					}
				}
			}
			return err2
		}
		break
	}

	err = fn(tx)
	if err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			glog.Errorf("Transaction rollback failed: %+v", err2)
		}
		return errors.Trace(err)
	}

	err = tx.Commit()
	if err != nil {
		return errors.Annotate(err, "commit transaction")
	}
	return nil
}

func initialize(db *sql.DB) error {
	err := tx(db, func(tx *sql.Tx) error {
		_, err := tx.Exec(`
        CREATE TABLE IF NOT EXISTS dfmigrate_state (
          param TEXT NOT NULL UNIQUE,
          value INTEGER NOT NULL
        )
      `)
		return err
	})
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func getCurrentMigrationID(db *sql.DB) (int, error) {
	curID := 0

	err := tx(db, func(tx *sql.Tx) error {
		err := tx.QueryRow(
			"SELECT value FROM dfmigrate_state WHERE param = $1", paramCurMigrationID,
		).Scan(&curID)
		if err != nil {
			if errors.Cause(err) == sql.ErrNoRows {
				curID = 0
			} else {
				return errors.Trace(err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, errors.Trace(err)
	}

	return curID, nil
}

func setCurrentMigrationID(db *sql.DB, curID int) error {
	err := tx(db, func(tx *sql.Tx) error {
		_, err := tx.Exec(`
      INSERT INTO dfmigrate_state (param, value) values ($1, $2)
      ON CONFLICT (param) DO UPDATE SET value = $2;
    `,
			paramCurMigrationID, curID,
		)
		if err != nil {
			return errors.Trace(err)
		}

		return nil
	})
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
