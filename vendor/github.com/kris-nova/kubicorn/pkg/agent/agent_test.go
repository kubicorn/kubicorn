// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"testing"
)

const (
	incorrectError  string = "error message incorrect\n got:       %v\n"
	expectedAnError string = "expected an error, received none."
)

func TestCheckKeyWithoutPassword(t *testing.T) {
	a := NewAgent()

	if a.CheckKey("./testdata/ssh_without_password.pub") == nil {
		t.Fatal(expectedAnError)
	}

	// clean test
	defer a.RemoveKeyUsingFile("./testdata/ssh_without_password.pub")
}

func TestAddKeyWithoutPassword(t *testing.T) {
	a := NewAgent()

	if a.CheckKey("./testdata/ssh_without_password.pub") == nil {
		t.Fatal(expectedAnError)
	}

	a, err := a.AddKey("./testdata/ssh_without_password.pub")
	if err != nil {
		t.Fatalf(incorrectError, err)
	}

	err = a.CheckKey("./testdata/ssh_without_password.pub")
	if err != nil {
		t.Fatalf(incorrectError, err)
	}

	// clean test
	defer a.RemoveKeyUsingFile("./testdata/ssh_without_password.pub")
}

func TestCheckKeyWithPassword(t *testing.T) {
	a := NewAgent()

	if a.CheckKey("./testdata/ssh_with_password.pub") == nil {
		t.Fatal(expectedAnError)
	}

	// clean test
	defer a.RemoveKeyUsingFile("./testdata/ssh_with_password.pub")
}

func TestAddKeyWithPassword(t *testing.T) {
	retriveSSHKeyPassword = func() ([]byte, error) {
		return []byte("kubicornbesttoolever"), nil
	}

	a := NewAgent()

	if a.CheckKey("./testdata/ssh_with_password.pub") == nil {
		t.Fatal(expectedAnError)
	}

	a, err := a.AddKey("./testdata/ssh_with_password.pub")
	if err != nil {
		t.Fatalf(incorrectError, err)
	}

	if err = a.CheckKey("./testdata/ssh_with_password.pub"); err != nil {
		t.Fatalf(incorrectError, err)
	}

	// clean test
	defer a.RemoveKeyUsingFile("./testdata/ssh_with_password.pub")
}

func TestAddKeyWithPasswordIncorrect(t *testing.T) {
	retriveSSHKeyPassword = func() ([]byte, error) {
		return []byte("random"), nil
	}

	a := NewAgent()

	if a.CheckKey("./testdata/ssh_with_password.pub") == nil {
		t.Fatal(expectedAnError)
	}

	if _, err := a.AddKey("./testdata/ssh_with_password.pub"); err == nil {
		t.Fatalf(expectedAnError)
	}

	if a.CheckKey("./testdata/ssh_with_password.pub") == nil {
		t.Fatal(expectedAnError)
	}
}

func TestRemoveKey(t *testing.T) {
	var err error
	a := NewAgent()

	// check that key doesnt exist
	if a.CheckKey("./testdata/ssh_without_password.pub") == nil {
		t.Fatal(expectedAnError)
	}

	if _, err = a.AddKey("./testdata/ssh_without_password.pub"); err != nil {
		t.Fatalf(incorrectError, err)
	}

	// should be able to remove this key
	if err = a.RemoveKeyUsingFile("./testdata/ssh_without_password.pub"); err != nil {
		t.Fatalf("no error expecting in removing key:\n        %v", err)
	}
}
