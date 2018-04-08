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

// Package ssh is used to connect to nodes over SSH.
package ssh

import (
	gossh "golang.org/x/crypto/ssh"
	"github.com/kubicorn/kubicorn/pkg/agent"
	"os"
	"fmt"
)

// SSHClient contains parameters for connection to the node.
type SSHClient struct {
	// IP address or FQDN of the node.
	Address string

	// Port of the node's SSH server.
	Port string

	// Path to the public key used for authentication.
	PubKeyPath string

	// Session is session created when connecting to the node.
	Session *gossh.Session

	// Client is basic Go SSH client needed to make SSH connection.
	Client *gossh.ClientConfig
}

// NewSSHClient returns a SSH client representation.
func NewSSHClient(address, port, username string) *SSHClient {
	s := &SSHClient{
		Address: address,
		Port: port,
		Session: nil,
		Client: &gossh.ClientConfig{
			User: username,
			Auth: []gossh.AuthMethod{},
			HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		},
	}

	// DEPRECATED: this approach is to be deprecated as we build the new SSH wrapper.
	sshAgent := agent.NewAgent()
	s.Client.Auth = append(s.Client.Auth, sshAgent.GetAgent())

	return s
}

// NewHeadlessConnection starts a headless connection against the node.
func (s *SSHClient) NewHeadlessConnection() error {
	conn, err := gossh.Dial("tcp", fmt.Sprintf("%s:%s", s.Address, s.Port), s.Client)
	if err != nil {
		return err
	}

	s.Session, err = conn.NewSession()
	return err
}

// NewTerminalConnection starts a terminal connection against the node.
func (s *SSHClient) NewTerminalConnection() error {
	if s.Session == nil {
		err := s.NewHeadlessConnection()
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	s.Session.Stdout = os.Stdout
	s.Session.Stderr = os.Stderr
	s.Session.Stdin = os.Stdin

	if err := s.Session.Shell(); err != nil {
		return err
	}

	err := s.Session.Wait()
	if _, ok := err.(*gossh.ExitError); ok {
		return nil
	}
	return err
}


func (s *SSHClient) Close() error {
	return s.Session.Close()
}
