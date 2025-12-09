package backend

import "embed"

// FrontendFiles contains the embedded frontend files
// Using all: prefix to include subdirectories like assets/
//
//go:embed all:frontend/dist
var FrontendFiles embed.FS
