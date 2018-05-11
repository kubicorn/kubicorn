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
	"fmt"
	"os"

	"strings"

	"github.com/kubicorn/kubicorn/pkg/local"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/ssh/auth"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// Client contains parameters for connection to the node.
type Client struct {
	// IP address or FQDN of the node.
	Address string

	// Port of the node's SSH server.
	Port string

	// ClientConfig is a basic Go SSH client needed to make SSH connection.
	// This is populated automatically from fields provided on Client creation time.
	ClientConfig *gossh.ClientConfig

	// Conn is connection to the remote SSH server.
	// Connection is made using the Connect function.
	Conn *gossh.Client
}

// NewClient returns a SSH client representation.
func NewClient(address, port, username, publicKeyPath string) *Client {
	s := &Client{
		Address: address,
		Port:    port,
		ClientConfig: &gossh.ClientConfig{
			User:            username,
			Auth:            []gossh.AuthMethod{},
			HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		},
		Conn: nil,
	}

	// If system agent is enabled in Kubicorn use it.
	sysAgent := false
	if os.Getenv("KUBICORN_SSH_AGENT") != "" {
		agent, err := auth.SystemAgentIfExists()
		if err != nil {
			logger.Warning("error initializing ssh agent: %v", err)
		}
		if agent != nil { // System agent exists.
			if err := auth.CheckKey(agent, local.Expand(publicKeyPath)); err == nil { // Key is found in agent.
				s.ClientConfig.Auth = append(s.ClientConfig.Auth, gossh.PublicKeysCallback(agent.Signers))
				sysAgent = true
			} else { // Key is not found in agent.
				logger.Critical("SSH key not present in system agent. Resorting to password authentication.")
				logger.Critical("To add SSH key to the agent, use the ssh-add command.")
			}
		} else {
			logger.Critical("SSH agent not found. Resorting to password authentication.")
		}
	}

	if !sysAgent { // Resort to password authentication if agent is not present.
		// Parse public and private key.
		privKeyPath := strings.Replace(local.Expand(publicKeyPath), ".pub", "", 1)
		k, err := auth.ParsePrivateKey(privKeyPath)
		if err != nil {
			logger.Critical("Unable to parse private key: %v", err)
		}

		signer, err := gossh.NewSignerFromKey(k)
		if err != nil {
			logger.Critical("Unable to parse private key: %v", err)
		}

		s.ClientConfig.Auth = append(s.ClientConfig.Auth, gossh.PublicKeys(signer))
	}

	return s
}

// Connect starts a headless connection against the node.
func (s *Client) Connect() error {
	conn, err := gossh.Dial("tcp", fmt.Sprintf("%s:%s", s.Address, s.Port), s.ClientConfig)
	if err != nil {
		return err
	}

	s.Conn = conn
	return nil
}

// StartInteractiveSession starts a terminal connection against the node.
func (s *Client) StartInteractiveSession() error {
	if s.Conn == nil {
		return fmt.Errorf("not connected to the server")
	}

	// Start interactive session.
	session, err := s.Conn.NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	// Bind session stdout, stderr, stdin to system's ones.
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	// It's required to bind to terminal, otherwise, session will fail with ioctl error.
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() {
		_ = terminal.Restore(fd, oldState)
	}()
	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		termWidth = 80
		termHeight = 24
	}
	modes := gossh.TerminalModes{
		gossh.ECHO: 1,
	}
	// TODO: this can be a bad approach, e.g. what if xterm is not available. Research more about this function.
	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return err
	}

	// Inform session if client terminal size changed.
	go func() {
		for {
			tw, th, _ := terminal.GetSize(fd)
			if termWidth != tw || termHeight != th {
				session.WindowChange(th, tw)
				termWidth = tw
				termHeight = th
			}
		}
	}()

	// Start shell session.
	if err := session.Shell(); err != nil {
		return err
	}

	// Wait for session to complete and check for error.
	return session.Wait()
}

// Execute executes command on the remote server and returns stdout and stderr output.
func (s *Client) Execute(cmd string) ([]byte, error) {
	if s.Conn == nil {
		return nil, fmt.Errorf("not connected to the server")
	}

	// Start interactive session.
	session, err := s.Conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = session.Close()
	}()

	return session.CombinedOutput(cmd)
}

// Close closes the SSH connection.
func (s *Client) Close() error {
	if s.Conn == nil {
		return fmt.Errorf("connection not existing")
	}
	return s.Conn.Close()
}
