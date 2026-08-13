//go:build doltlite

package db

import (
	"database/sql"
	"sync"

	sqlite3 "github.com/mattn/go-sqlite3"
)

var registerDoltLiteDriverOnce sync.Once

func registerDoltLiteDriver(_ Config) error {
	registerDoltLiteDriverOnce.Do(func() {
		sql.Register(doltLiteDriverName, &sqlite3.SQLiteDriver{})
	})
	return nil
}
