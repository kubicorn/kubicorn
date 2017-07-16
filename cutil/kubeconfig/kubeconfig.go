package kubeconfig

import (
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/local"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func GetConfig(existing *cluster.Cluster) error {
	user := existing.Ssh.User
	pubKeyPath := local.Expand(existing.Ssh.PublicKeyPath)
	privKeyPath := strings.Replace(pubKeyPath, ".pub", "", 1)
	address := fmt.Sprintf("%s:%s", existing.KubernetesApi.Endpoint, "22")
	remotePath := fmt.Sprintf("/home/%s/.kube/config", user)
	localPath := fmt.Sprintf("%s/.kube/config", local.Home())

	//fmt.Println(pubKeyPath)
	//fmt.Println(privKeyPath)
	//fmt.Println(address)
	//fmt.Println(user)
	//fmt.Println(remotePath)
	//fmt.Println(localPath)

	pemBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return err
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return err
	}

	auths := []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}

	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            auths,
	}
	sshConfig.SetDefaults()

	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	c, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer c.Close()
	r, err := c.Open(remotePath)
	if err != nil {
		return err
	}
	defer r.Close()
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

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
	RetryAttempts     = 40
	RetrySleepSeconds = 3
)

func RetryGetConfig(existing *cluster.Cluster) error {
	for i := 0; i <= RetryAttempts; i++ {
		err := GetConfig(existing)
		if err != nil {
			if strings.Contains(err.Error(), "file does not exist") {
				logger.Debug("Waiting for Kubernetes to come up..")
				time.Sleep(time.Duration(RetrySleepSeconds) * time.Second)
				continue
			}
			return nil
		}
		break
	}
	return nil
}
