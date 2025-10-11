package buildinfo

// These are defaulted for local dev; CI overwrites them via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
