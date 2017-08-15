package jsonfs

import (
	"testing"
	"github.com/kris-nova/kubicorn/profiles"
	"github.com/kris-nova/kubicorn/state"
	"reflect"
)

func TestJsonFileSystem(t *testing.T) {
	c := profiles.NewSimpleAmazonCluster("jsonfs-test")
	o := &JSONFileSystemStoreOptions{
		AbsolutePath: ".test/",
		ClusterName: c.Name,
	}
	fs := NewJSONFileSystemStore(o)
	if err := fs.Destroy(); err != nil {
		t.Fatalf("Error destroying any existing state: %v", err)
	}
	if fs.Exists() {
		t.Fatalf("State shouldn't exist because we just destroyed it, but Exists() returned true")
	}
	if err := fs.Commit(c); err != nil {
		t.Fatalf("Error committing cluster: %v", err)
	}
	files, err := fs.List()
	if err != nil {
		t.Fatalf("Error listing files: %v", err)
	}
	if len(files) < 1 {
		t.Fatalf( "Expected at least one cluster, got: %v", len(files))
	}
	if files[0] != state.ClusterJsonFile {
		t.Fatalf("Expected file name to be %v, got %v", state.ClusterJsonFile, files[0])
	}
	read, err := fs.GetCluster()
	if err != nil {
		t.Fatalf("Error getting cluster: %v", err)
	}
	if !reflect.DeepEqual(read, c){
		t.Fatalf("Cluster in doesn't equal cluster out")
	}
	if err = fs.Destroy(); err != nil {
		t.Fatalf("Error cleaning up state: %v", err)
	}
}
