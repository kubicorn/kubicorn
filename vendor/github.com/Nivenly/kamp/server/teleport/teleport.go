package teleport

import (
	"fmt"
	"github.com/Nivenly/kamp/local"
	"github.com/Nivenly/kamp/server"
	"github.com/gravitational/teleport/tool/teleport/common"
	"os"
)

type TeleportServer struct {
}

func NewServer() server.Server {
	return &TeleportServer{}
}

func (t *TeleportServer) Authorize(auth *server.RsaAuth) error {
	local.Info("Authenticating with Teleport [%s]", auth.Username)
	local.Debug("RSA Key bytes: %d", len(auth.PublicKey))
	return nil
}

func (t *TeleportServer) Run() error {
	uid := os.Geteuid()
	if uid != 0 {
		return fmt.Errorf("Must run server as root, currently running as [%s-%d]", local.User(), uid)
	}
	local.Info("Running Teleport SSH Server")
	args := []string{"start"}
	common.Run(args, false)
	return nil
}
