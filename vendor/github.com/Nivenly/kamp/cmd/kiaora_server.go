// Copyright © 2017 The Kamp Authors
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

package cmd

import (
	"github.com/Nivenly/kamp/local"
	server2 "github.com/Nivenly/kamp/server"
	"github.com/Nivenly/kamp/server/teleport"
	"github.com/spf13/cobra"
	"io/ioutil"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "The Kiaora server for mounting remote file systems.",
	Long: `Run Kiaora as a server that will listen for connections.
This may be run within the context of a pod, or within the context of a user's local workstation.'`,
	Run: func(cmd *cobra.Command, args []string) {
		local.LogLevel = O.Verbosity
		serverOpt.PublicKeyPath = local.Expand(serverOpt.PublicKeyPath)
		data, err := ioutil.ReadFile(serverOpt.PublicKeyPath)
		Check(err)
		serverOpt.PublicKeyData = data
		err = RunServer(serverOpt)
		Check(err)
	},
}

func init() {
	kiaoraCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverOpt.PublicKeyPath, "identity-file", "i", "~/.ssh/id_rsa.pub", "The public key to authorize with the server. A user with the private key that corresponds to the public key will be able to authenticate with the server.")
	serverCmd.Flags().StringVarP(&serverOpt.User, "user", "u", local.User(), "The user that will be able to access the server with the provided identity file. The server will automatically handle provisioning the user and the authenticated public key.")
}

var (
	serverOpt = &ServerOptions{}
)

type ServerOptions struct {
	PublicKeyPath string
	PublicKeyData []byte
	User          string
}

func RunServer(options *ServerOptions) error {

	// ------------------------ Run SSH server ------------------------
	server := teleport.NewServer()
	server.Authorize(&server2.RsaAuth{
		PublicKey: options.PublicKeyData,
		Username:  options.User,
	})
	err := server.Run()
	if err != nil {
		return err
	}
	return nil
}
