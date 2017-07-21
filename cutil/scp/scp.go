package scp

import (
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/local"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type SecureCopier struct {
	RemoteUser     string
	RemoteAddress  string
	RemotePort     string
	PrivateKeyPath string
}

func NewSecureCopier() *SecureCopier {
	return &SecureCopier{}
}

func (s *SecureCopier) Read(remotePath, localPath string) ([]byte, error) {
	pemBytes, err := ioutil.ReadFile(s.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	signer, err := GetSigner(pemBytes)
	if err != nil {
		return nil, err
	}
	auths := []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.RemoteUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            auths,
	}
	sshConfig.SetDefaults()
	conn, err := ssh.Dial("tcp", s.RemoteAddress, sshConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c, err := sftp.NewClient(conn)
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

func (s *SecureCopier) Write(localPath, remotePath string) error {
	logger.Critical("Write not yet implemented!")
}

func GetConfig(options *ScpOptions) error {
	user := existing.Ssh.User
	pubKeyPath := local.Expand(existing.Ssh.PublicKeyPath)
	privKeyPath := strings.Replace(pubKeyPath, ".pub", "", 1)
	address := fmt.Sprintf("%s:%s", existing.KubernetesApi.Endpoint, "22")
	localPath := fmt.Sprintf("%s/.kube/config", local.Home())
	remotePath := ""
	if user == "root" {
		remotePath = "/root/.kube/config"
	} else {
		remotePath = fmt.Sprintf("/home/%s/.kube/config", user)
	}

	//fmt.Println(pubKeyPath)
	//fmt.Println(privKeyPath)
	//fmt.Println(address)
	//fmt.Println(user)
	//fmt.Println(remotePath)
	//fmt.Println(localPath)








	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		empty := []byte("")
		err := ioutil.WriteFile(localPath, empty, 0755)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(localPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	_, err = f.WriteString(string(bytes))
	if err != nil {
		return err
	}
	defer f.Close()
	logger.Always("Wrote kubeconfig to [%s]", localPath)
	return nil
}

const (
	RetryAttempts     = 90
	RetrySleepSeconds = 3
)

func RetryGetConfig(existing *cluster.Cluster) error {
	for i := 0; i <= RetryAttempts; i++ {
		err := GetConfig(existing)
		if err != nil {
			if strings.Contains(err.Error(), "file does not exist") || strings.Contains(err.Error(), "getsockopt: connection refused") {
				//logger.Warning(err.Error())
				logger.Debug("Waiting for Kubernetes to come up..")
				time.Sleep(time.Duration(RetrySleepSeconds) * time.Second)
				continue
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("Timedout writing kubeconfig")
}

func GetSigner(pemBytes []byte) (ssh.Signer, error) {
	signerwithoutpassphrase, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		fmt.Print("SSH Key Passphrase [none]: ")
		passPhrase, err := terminal.ReadPassword(0)
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
