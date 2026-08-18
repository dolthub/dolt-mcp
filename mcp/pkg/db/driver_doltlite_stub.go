//go:build !doltlite

package db

import "errors"

// ErrDoltLiteNotSupported is returned when the DoltLite dialect is requested
// from a binary compiled without the "doltlite" build tag.
var ErrDoltLiteNotSupported = errors.New(
	"this binary was built without DoltLite support; rebuild with -tags \"doltlite libsqlite3\" and link against libdoltlite (see README)",
)

func registerDoltLiteDriver(_ Config) error {
	return ErrDoltLiteNotSupported
}
