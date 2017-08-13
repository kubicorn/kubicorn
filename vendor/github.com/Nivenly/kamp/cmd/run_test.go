package cmd

import "testing"

func TestRun(t *testing.T) {
	o := &RunOptions{}
	err := RunRun(o)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
