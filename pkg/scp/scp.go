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

package scp

import (
	"fmt"
	"io/ioutil"

	"github.com/kubicorn/kubicorn/pkg/ssh"
	"github.com/pkg/sftp"
)

// Client represents SCP client.
type Client struct {
	client *ssh.Client
}

func NewSCPClient(sshClient *ssh.Client) *Client {
	return &Client{
		client: sshClient,
	}
}

// ReadBytes reads from remote location.
func (cl *Client) ReadBytes(remotePath string) ([]byte, error) {
	if cl.client.Conn == nil {
		return nil, fmt.Errorf("Connection not established.")
	}

	c, err := sftp.NewClient(cl.client.Conn)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	r, err := c.Open(remotePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// WriteBytes writes to remote location.
func (cl *Client) WriteBytes(remotePath string, content []byte) error {
	if cl.client.Conn == nil {
		return fmt.Errorf("Connection not established.")
	}

	c, err := sftp.NewClient(cl.client.Conn)
	if err != nil {
		return err
	}
	defer c.Close()

	f, err := c.Create(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	if err != nil {
		return err
	}

	return nil
}
