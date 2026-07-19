// Package static embeds the compiled Tailwind CSS output and the vendored
// Basecoat CSS/JS bundle directly into the server binary, per the project's
// single-static-binary goal (docs/TRD.md). Both files are produced by
// `make css` / `make basecoat` before the Go build runs (see Makefile and
// .air.toml's pre_cmd) — this file only wires them into the binary.
package static

import "embed"

// FS contains web/static/css/output.css and the Basecoat bundle. Neither
// file is committed to the repo (see .gitignore); they must exist on disk
// at build time, which `make dev` / `make build` guarantee via their
// prerequisite targets.
//
//go:embed css/output.css vendor/basecoat/basecoat.min.css vendor/basecoat/basecoat.min.js
var FS embed.FS
