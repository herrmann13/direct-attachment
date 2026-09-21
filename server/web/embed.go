// Package web embeds the static assets of the mobile capture page so the
// server ships as a single binary.
package web

import "embed"

// FS contains the static assets of the mobile capture page.
//
//go:embed index.html style.css app.js tweetnacl.min.js
var FS embed.FS
