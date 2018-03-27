package version

import (
	"encoding/json"
	"runtime"
	"time"

	"github.com/kubicorn/kubicorn/pkg/logger"
)

var (
	versionFile = "/src/github.com/kubicorn/kubicorn/VERSION"

	// The GitSha of the current commit (automatically set at compile time)
	GitSha string

	// The Version of the program from the VERSION file (automatically set at compile time)
	kubicornVersion string
)

// Version represents Kubicorn version.
type Version struct {
	Version string
	GitCommit string
	BuildDate string
	GoVersion string
	GOOS string
	GOArch string
}

// GetVersion returns Kubicorn version.
func GetVersion() *Version {
	return &Version {
		Version: kubicornVersion,
		GitCommit: GitSha,
		BuildDate: time.Now().UTC().String(),
		GoVersion: runtime.Version(),
		GOOS: runtime.GOOS,
		GOArch: runtime.GOARCH,
	}
}

// GetVersionJSON returns Kubicorn version in JSON format.
func GetVersionJSON() string {
	verBytes, err := json.Marshal(GetVersion())
	if err != nil {
		logger.Critical("Unable to marshal version struct: %v", err)
	}
	return string(verBytes)
}