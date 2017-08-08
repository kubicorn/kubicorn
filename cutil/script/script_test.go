package script

import "testing"

func TestBuildBootstrapScriptHappy(t *testing.T) {
	scripts := []string{
		"vpn/meshbirdMaster.sh",
		"digitalocean_k8s_ubuntu_16.04_master.sh",
	};
	_, err := BuildBootstrapScript(scripts)
	if err != nil {
		t.Fatalf("Unable to get scripts: %v", err)
	}
}

func TestBuildBootstrapScriptSad(t *testing.T) {
	scripts := []string{
		"vpn/meshbirdMaster.s",
		"digitalocean_k8s_ubuntu_16.04_master.s",
	};
	_, err := BuildBootstrapScript(scripts)
	if err == nil {
		t.Fatalf("Merging non existing scripts: %v", err)
	}
}