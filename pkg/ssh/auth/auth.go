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

package auth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// retriveSSHKeyPassword takes password from terminal.
var retriveSSHKeyPassword = func() ([]byte, error) {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return nil, fmt.Errorf("cannot detect terminal")
	}

	fmt.Print("SSH Key Passphrase: ")
	passPhrase, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}

	fmt.Println("")
	return passPhrase, nil
}

// ParsePrivateKey unlocks and parses private key.
func ParsePrivateKey(path string) (interface{}, error) {
	// Read key form file.
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, err
	}

	var priv interface{}
	if x509.IsEncryptedPEMBlock(block) {
		// Obtain password from terminal.
		password, err := retriveSSHKeyPassword()
		if err != nil {
			return nil, err
		}

		// Decrypt the PEM block.
		b, err := x509.DecryptPEMBlock(block, []byte(password))
		if err != nil {
			return nil, err
		}

		// Parse private key.
		priv, err = x509.ParsePKCS1PrivateKey(b)
		if err != nil {
			return nil, err
		}
	} else { // If key is not encrypted, just parse it as it is.
		priv, err = ssh.ParseRawPrivateKey(b)
		if err != nil {
			return nil, err
		}
	}

	return priv, nil
}
