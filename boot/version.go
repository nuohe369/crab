package boot

// Version information, can be overridden at build time using ldflags:
// go build -ldflags "-X github.com/nuohe369/crab/boot.Version=1.0.0 -X github.com/nuohe369/crab/boot.GitCommit=$(git rev-parse --short HEAD) -X github.com/nuohe369/crab/boot.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// VersionInfo returns formatted version information.
func VersionInfo() string {
	return "crab v" + Version + " (commit: " + GitCommit + ", built: " + BuildTime + ")"
}

// VersionShort returns short version string.
func VersionShort() string {
	return "crab v" + Version
}
