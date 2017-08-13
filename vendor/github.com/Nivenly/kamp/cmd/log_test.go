package cmd

import "testing"

func TestLog(t *testing.T) {
	o := &LogOptions{}
	err := RunLog(o)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
