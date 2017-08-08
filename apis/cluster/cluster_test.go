package cluster

import "testing"

func TestNewClusterHappy(t *testing.T) {
	result := NewCluster("AwesomeCluster");
	if result.Name != "AwesomeCluster" {
		t.Errorf("Name for new cluster not set. Should be AwesomeCluster got : %s", result.Name)
	}
}
