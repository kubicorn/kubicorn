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

package sshcmd

import (
	"fmt"
	"os"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/agent"
	"github.com/kris-nova/kubicorn/cutil/local"
	"golang.org/x/crypto/ssh"
)

func ExecCommandSSH(existing *cluster.Cluster, sshAgent *agent.Keyring, command string) error {
	user := existing.SSH.User
	address := fmt.Sprintf("%s:%s", existing.KubernetesAPI.Endpoint, existing.SSH.Port)
	pubKeyPath := local.Expand(existing.SSH.PublicKeyPath)
	if existing.SSH.Port == "" {
		existing.SSH.Port = "22"
	}
	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Check for key
	if err := sshAgent.CheckKey(pubKeyPath); err != nil {
		if keyring, err := sshAgent.AddKey(pubKeyPath); err != nil {
			return err
		} else {
			sshAgent = keyring
		}
	}

	if sshAgent != nil && os.Getenv("KUBICORN_FORCE_DISABLE_SSH_AGENT") == "" {
		sshConfig.Auth = append(sshConfig.Auth, sshAgent.GetAgent())
	}

	sshConfig.SetDefaults()
	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	session.Stdout = os.Stdout
	session.Stdin = os.Stdin
	session.Stderr = os.Stderr

	return session.Run(command)
}
