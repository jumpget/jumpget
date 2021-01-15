package ssh

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"time"
)

type RemoteExecutor struct {
	user           string
	host           string
	port           int
	privateKeyFile string
	client         *ssh.Client
}

func (exec *RemoteExecutor) Close() (err error) {
	return exec.client.Close()
}

func NewRemoteExecutor(keyFile string, user string, host string, port int) *RemoteExecutor {
	exec := &RemoteExecutor{
		privateKeyFile: keyFile,
		user:           user,
		host:           host,
		port:           port,
	}
	return exec
}

func (exec *RemoteExecutor) Execute(command string) ([]byte, error) {
	session, err := exec.client.NewSession()
	if err != nil {
		exec.client.Close()
		return nil, err
	}
	defer session.Close()
	return session.CombinedOutput(command)
}

func paresPrivateKey(keyfile string) (key ssh.Signer, err error) {
	buf, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to parse the private key: %v", err))
	}
	return key, nil
}

func (exec *RemoteExecutor) Init() error {
	signer, err := paresPrivateKey(exec.privateKeyFile)
	if err != nil {
		return err
	}
	publicKeys := ssh.PublicKeys(signer)
	sshConfig := &ssh.ClientConfig{
		Timeout: time.Second * 10,
		User:    exec.user,
		Auth:    []ssh.AuthMethod{publicKeys},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	exec.client, err = ssh.Dial(
		"tcp",
		fmt.Sprintf("%v:%v", exec.host, exec.port),
		sshConfig)
	if err != nil {
		return err
	}
	fmt.Printf("connected to host: %v:%v\n", exec.host, exec.port)
	return nil
}
