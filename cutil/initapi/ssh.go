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

package initapi

import (
	"golang.org/x/crypto/ssh/terminal"
	"github.com/kris-nova/klone/pkg/local"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"io/ioutil"
	"strings"
	"crypto/md5"
	"fmt"
	"golang.org/x/crypto/ssh"
	"github.com/gravitational/trace"
	"syscall"
)

func sshLoader(initCluster *cluster.Cluster) (*cluster.Cluster, error) {
	if initCluster.Ssh.PublicKeyPath != "" {
		bytes, err := ioutil.ReadFile(local.Expand(initCluster.Ssh.PublicKeyPath))
		if err != nil {
			return nil, err
		}
		initCluster.Ssh.PublicKeyData = bytes
		privateBytes, err := ioutil.ReadFile(strings.Replace(local.Expand(initCluster.Ssh.PublicKeyPath), ".pub", "", 1))
		if err != nil {
			return nil, err
		}
		fp, err := PrivateKeyFingerprint(privateBytes)
		if err != nil {
			return nil, err
		}
		initCluster.Ssh.PublicKeyFingerprint = fp
	}

	return initCluster, nil
}

func fingerprint(key ssh.PublicKey) string {
	sum := md5.Sum(key.Marshal())
	parts := make([]string, len(sum))
	for i := 0; i < len(sum); i++ {
		parts[i] = fmt.Sprintf("%0.2x", sum[i])
	}
	return strings.Join(parts, ":")
}

func AuthorizedKeyFingerprint(publicKey []byte) (string, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey(publicKey)
	if err != nil {
		return "", err
	}

	return fingerprint(key), nil
}

func PrivateKeyFingerprint(keyBytes []byte) (string, error) {
	signer, err := GetSigner(keyBytes)
	if err != nil {
		return "", trace.Wrap(err)
	}
	return fingerprint(signer.PublicKey()), nil
}

func GetSigner(pemBytes []byte) (ssh.Signer, error) {
	signerwithoutpassphrase, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		fmt.Print("SSH Key Passphrase [none]: ")
		passPhrase, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			return nil, err
		}
		signerwithpassphrase, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, passPhrase)
		if err != nil {
			return nil, err
		} else {
			return signerwithpassphrase, err
		}
	} else {
		return signerwithoutpassphrase, err
	}
}