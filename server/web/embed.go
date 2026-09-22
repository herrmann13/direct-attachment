// Package web embeds the static assets of the mobile capture page so the
// server ships as a single binary.
package web

import "embed"

// FS contains the static assets of the mobile capture page.
//
//go:embed index.html privacy.html style.css app.js tweetnacl.min.js cropper.min.js cropper.min.css
var FS embed.FS
