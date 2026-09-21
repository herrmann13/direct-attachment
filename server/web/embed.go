// Package web embeds the static assets of the mobile capture page so the
// server ships as a single binary.
package web

import "embed"

// FS contains index.html, style.css and app.js.
//
//go:embed index.html style.css app.js
var FS embed.FS
