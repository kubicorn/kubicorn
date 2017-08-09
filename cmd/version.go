package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Verify Kubicorn version",
	Long: `Use this command to check the version of Kubicorn.

This command will return the version of the Kubicorn binary.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunVersion(vo)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

// VersionOptions contains fields for version output
type VersionOptions struct {
	Version   string `json:"Version"`
	GitCommit string `json:"GitCommit"`
	BuildDate string `json:"BuildDate"`
	GOVersion string `json:"GOVersion"`
	GOARCH    string `json:"GOARCH"`
	GOOS      string `json:"GOOS"`
}

var vo = &VersionOptions{}

func init() {
	RootCmd.AddCommand(versionCmd)
}

// RunVersion populates VersionOptions and prints to stdout
func RunVersion(vo *VersionOptions) error {

	vo.Version = getVersion()
	vo.GitCommit = getGitCommit()
	vo.BuildDate = time.Now().UTC().String()
	vo.GOVersion = runtime.Version()
	vo.GOARCH = runtime.GOARCH
	vo.GOOS = runtime.GOOS
	voBytes, err := json.Marshal(vo)
	if err != nil {
		return err
	}
	fmt.Println("Kubicorn version: ", string(voBytes))
	return nil
}

var (
	versionFile = "/src/github.com/kris-nova/kubicorn/VERSION"
)

func getVersion() string {
	path := filepath.Join(os.Getenv("GOPATH") + versionFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return string(vBytes)
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// ignore error
		return ""
	}
	return string(output)
}
