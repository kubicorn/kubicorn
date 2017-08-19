package cmd

import (
	"fmt"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
)

type EditOptions struct {
	Options
	Editor string
}

var eo = &EditOptions{}

var editCmd = &cobra.Command{
	Use:   "edit <NAME>",
	Short: "Edit a cluster state",
	Long:  `Use this command to edit a state.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			logger.Critical("Too many arguments.")
			os.Exit(1)
		} else {
			eo.Name = args[0]
		}

		err := RunEdit(eo)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

func init() {
	editCmd.Flags().StringVarP(&eo.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	editCmd.Flags().StringVarP(&eo.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	editCmd.Flags().StringVarP(&eo.Editor, "editor", "e", strEnvDef("KUBICORN_DEFAULT_EDITOR", "vi"), "The editor used to edit the state store")

	RootCmd.AddCommand(editCmd)
}

func RunEdit(options *EditOptions) error {
	options.StateStorePath = expandPath(options.StateStorePath)

	name := options.Name
	// Register state store
	var stateStore state.ClusterStorer
	switch options.StateStore {
	case "fs":
		logger.Info("Selected [fs] state store")
		stateStore = fs.NewFileSystemStore(&fs.FileSystemStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: name,
		})
	}

	// Check if state store exists
	if !stateStore.Exists() {
		return fmt.Errorf("State store [%s] does not exists, can't edit", name)
	}
	stateContent, err := stateStore.ReadStore()
	if err != nil {
		return err
	}

	fpath := os.TempDir() + "/kubicorn_cluster.tmp"
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	ioutil.WriteFile(fpath, stateContent, 0664)
	f.Close()

	path, err := exec.LookPath(options.Editor)
	if err != nil {
		os.Remove(fpath)
		return err
	}

	cmd := exec.Command(path, fpath)
	err = cmd.Start()
	if err != nil {
		os.Remove(fpath)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logger.Debug("Error while editing. Error: %v", err)
		os.Remove(fpath)
		return err
	} else {
		logger.Info("Successfull edit")
	}

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		os.Remove(fpath)
		return err
	}

	cluster, err := stateStore.BytesToCluster(data)
	if err != nil {
		os.Remove(fpath)
		return err
	}

	// Init new state store with the cluster resource
	err = stateStore.Commit(cluster)
	if err != nil {
		os.Remove(fpath)
		return fmt.Errorf("Unable to init state store: %v", err)
	}
	os.Remove(fpath)

	return nil
}
