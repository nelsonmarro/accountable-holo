// Package migrations represent de db migrations
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
