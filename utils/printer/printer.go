package printer

import (
	"github.com/pingcap/log"
	"go.uber.org/zap"

	_ "runtime" // import link package
	_ "unsafe"  // required by go:linkname
)

// Version information.
var (
	AutoIndexBuildTS   = "None"
	AutoIndexGitHash   = "None"
	AutoIndexGitBranch = "None"
)

//go:linkname buildVersion runtime.buildVersion
var buildVersion string

// PrintAutoIndexInfo prints the AutoIndex version information.
func PrintAutoIndexInfo() {
	log.Info("Welcome to auto-index",
		zap.String("Git Commit Hash", AutoIndexGitHash),
		zap.String("Git Branch", AutoIndexGitBranch),
		zap.String("UTC Build Time", AutoIndexBuildTS),
		zap.String("GoVersion", buildVersion))
}
