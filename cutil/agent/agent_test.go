package agent

import "testing"

func TestCheckKeyWithoutPassword(t *testing.T) {
	a := NewAgent()

	err := a.CheckKey("./testdata/ssh_without_password.pub")
	if err == nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}
}

func TestAddKeyWithoutPassword(t *testing.T) {
	a := NewAgent()

	err := a.CheckKey("./testdata/ssh_without_password.pub")
	if err == nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}

	a, err = a.AddKey("./testdata/ssh_without_password.pub")
	if err != nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}

	err = a.CheckKey("./testdata/ssh_without_password.pub")
	if err != nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}
}

func TestCheckKeyWithPassword(t *testing.T) {
	a := NewAgent()

	err := a.CheckKey("./testdata/ssh_with_password.pub")
	if err == nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}
}

func TestAddKeyWithPassword(t *testing.T) {
	retriveSSHKeyPassword = func() ([]byte, error) {
		return []byte("kubicornbesttoolever"), nil
	}
	a := NewAgent()

	err := a.CheckKey("./testdata/ssh_with_password.pub")
	if err == nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}

	a, err = a.AddKey("./testdata/ssh_with_password.pub")
	if err != nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}

	err = a.CheckKey("./testdata/ssh_with_password.pub")
	if err != nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}
}

func TestAddKeyWithPasswordIncorrect(t *testing.T) {
	retriveSSHKeyPassword = func() ([]byte, error) {
		return []byte("random"), nil
	}
	a := NewAgent()

	err := a.CheckKey("./testdata/ssh_with_password.pub")
	if err == nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}

	a, err = a.AddKey("./testdata/ssh_with_password.pub")
	if err == nil {
		t.Fatalf("error message incorrect\n"+
			"got:       %v\n", err)
	}
}
