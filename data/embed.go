// Package data embeds the static trust material served by the trust portal.
package data

import "embed"

// FS contains all static trust material files embedded at compile time.
//
//go:embed *.html *.svg *.json certs/*.crt keys/*.pub
var FS embed.FS
