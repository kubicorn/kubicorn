package cmd

import "testing"

func TestStatus(t *testing.T) {
	o := &StatusOptions{}
	err := RunStatus(o)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
