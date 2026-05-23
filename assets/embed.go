// Package assets exposes the embedded static files (CSS, JS, templates).
package assets

import "embed"

//go:embed static/* templates/*
var FS embed.FS
