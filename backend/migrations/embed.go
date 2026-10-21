// Package migrations embeds the goose SQL migrations so the API image can run
// them without a separately installed CLI.
package migrations

import "embed"

// FS contains every migration file in this directory.
//
//go:embed *.sql
var FS embed.FS
