//go:build doltlite

package db

import (
	"database/sql"
	"sync"

	sqlite3 "github.com/mattn/go-sqlite3"
)

// DoltLite is consumed through the stock SQLite C API: the server links
// against libdoltlite (a SQLite fork) via mattn/go-sqlite3 built with the
// libsqlite3 tag. Building this file requires cgo and libdoltlite, e.g.:
//
//	CGO_CFLAGS="-I/path/to/doltlite/build" \
//	CGO_LDFLAGS="/path/to/doltlite/build/libdoltlite.a -lz -lpthread" \
//	go build -tags "doltlite libsqlite3" ./mcp/cmd/dolt-mcp-server

var registerDoltLiteDriverOnce sync.Once

// registerDoltLiteDriver registers the "doltlite" database/sql driver. The
// Author configuration is replayed by newDoltLiteTransaction rather than
// captured here: database/sql driver registration is process-global, while
// different embedded database handles may use different authors.
func registerDoltLiteDriver(_ Config) error {
	registerDoltLiteDriverOnce.Do(func() {
		sql.Register(doltLiteDriverName, &sqlite3.SQLiteDriver{})
	})
	return nil
}
