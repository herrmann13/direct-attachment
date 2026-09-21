package upload

import "net/http"

// detectContentType wraps http.DetectContentType so it can be overridden in
// tests without touching the production validator. It sniffs the first 512
// bytes of data to determine the real media type.
var detectContentType = func(data []byte) string {
	return http.DetectContentType(data)
}
