package snap

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/mholt/archiver"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Snapshot struct {
	infra        *cluster.Cluster
	app          []byte
	absolutePath string
}

const (
	InfrastructureRepresentationName = "infra.k8s"
	ApplicationRepresentationName    = "apps.k8s"
)

func NewSnapshot(actualCluster *cluster.Cluster, appData []byte, absolutePath string) *Snapshot {
	return &Snapshot{
		infra:        actualCluster,
		app:          appData,
		absolutePath: absolutePath,
	}
}

func (s *Snapshot) AbsolutePath() string {
	return s.absolutePath
}

func (s *Snapshot) Bytes() []byte {
	return s.app
}

func (s *Snapshot) WriteCompressedFile() error {
	s.EnsureSnapshotName()
	dir, err := ioutil.TempDir("", "kubicorn")
	if err != nil {
		return err
	}

	// Write Infra
	infra := filepath.Join(dir, InfrastructureRepresentationName)
	infraBytes, err := yaml.Marshal(s.infra)
	if err != nil {
		return err
	}
	ioutil.WriteFile(infra, infraBytes, 0755)

	// Write App
	app := filepath.Join(dir, ApplicationRepresentationName)
	ioutil.WriteFile(app, s.app, 0755)

	logger.Debug("Snapshot path: %s", s.absolutePath)
	err = archiver.TarGz.Make(s.absolutePath, []string{dir})
	if err != nil {
		return err
	}
	return nil
}

func (s *Snapshot) EnsureSnapshotName() {
	if s.absolutePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			logger.Critical("Unable to get current working directory: %v", err)
			return
		}
		timestamp := time.Now().Format(time.RFC3339)
		s.absolutePath = fmt.Sprintf("%s/%s-%s.k8s", wd, s.infra.Name, timestamp)
	}
}
