package utils

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

func SshCommand(keyFile string, user string, host string, port int, command string) string {
	client, session, err := connectToHost(keyFile, user, fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		panic(err)
	}
	out, err := session.CombinedOutput(command)
	if err != nil {
		panic(err)
	}
	client.Close()
	return string(out)
}

func getKeyFile(keyfile string) (key ssh.Signer) {
	buf, err := ioutil.ReadFile(keyfile)
	if err != nil {
		panic(err)
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		panic(err)
	}
	return key
}

func connectToHost(keyfile, user, host string) (*ssh.Client, *ssh.Session, error) {
	publicKeys := ssh.PublicKeys(getKeyFile(keyfile))
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{publicKeys},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}
