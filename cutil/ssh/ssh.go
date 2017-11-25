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

package ssh

import (
	"bytes"
	"fmt"
	"time"

	"github.com/kris-nova/kubicorn/cutil/agent"
	"golang.org/x/crypto/ssh"
)

type SSHDetails struct {
	RemoteUser     string
	RemoteAddress  string
	RemotePort     string
	PrivateKeyPath string
	SSHAgent       *agent.Keyring
}

func NewSSHDetails(remoteUser, remoteAddress, remotePort, privateKeyPath string, sshAgent *agent.Keyring) *SSHDetails {
	return &SSHDetails{
		RemoteUser:     remoteUser,
		RemoteAddress:  remoteAddress,
		RemotePort:     remotePort,
		PrivateKeyPath: privateKeyPath,
		SSHAgent:       sshAgent,
	}
}

func (s *SSHDetails) ExecuteCommand(command string) ([]byte, error) {
	sshConfig := &ssh.ClientConfig{
		User:            s.RemoteUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(3 * time.Second),
	}

	// Check for key
	if err := s.SSHAgent.CheckKey(s.PrivateKeyPath + ".pub"); err != nil {
		if keyring, err := s.SSHAgent.AddKey(s.PrivateKeyPath + ".pub"); err != nil {
			return nil, err
		} else {
			s.SSHAgent = keyring
		}
	}

	sshConfig.Auth = append(sshConfig.Auth, s.SSHAgent.GetAgent())

	sshConfig.SetDefaults()
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", s.RemoteAddress, s.RemotePort), sshConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = session.Close()
	}()

	var buf bytes.Buffer
	session.Stdout = &buf

	if err := session.Run(command); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
