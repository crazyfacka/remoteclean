package modules

import (
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

// GetSSHConn returns a valid SSH connection
func GetSSHConn(confs map[string]interface{}) (*ssh.Client, error) {
	host := confs["host"].(string)
	port := int(confs["port"].(float64))
	user := confs["user"].(string)
	key := confs["key"].(string)

	homeDirectory, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	hostKeyCallback, err := knownhosts.New(filepath.Join(homeDirectory, ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			publicKeyFile(key),
		},
		HostKeyCallback: hostKeyCallback,
	}

	conn, err := ssh.Dial("tcp", host+":"+strconv.Itoa(port), sshConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
