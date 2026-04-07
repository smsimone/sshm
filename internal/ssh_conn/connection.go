package sshconn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/ssh"
)

type PrivateKey struct {
	Content    string  `json:"content"`
	Passphrase *string `json:"passphrase"`
}

type Connection struct {
	Label    string      `json:"label"`
	Host     string      `json:"host"`
	Port     int         `json:"port"`
	Username string      `json:"username"`
	Password *string     `json:"password,omitempty"`
	PvtKey   *PrivateKey `json:"private_key,omitempty"`
}

func (c *Connection) String() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%s:%s@%s:%d", c.Username, "none", c.Host, c.Port)
	}
	return string(bytes)
}

func (c *Connection) AuthMethods() []ssh.AuthMethod {
	methods := []ssh.AuthMethod{}
	if c.Password != nil {
		methods = append(methods, ssh.Password(*c.Password))
	}

	if c.PvtKey != nil {
		methods = append(methods, ssh.PublicKeys(c.PvtKey.signer()))
	}

	return methods
}

func (pvt *PrivateKey) signer() ssh.Signer {
	content, err := base64.StdEncoding.DecodeString(pvt.Content)
	if err != nil {
		panic(fmt.Sprintf("Failed to read private key content: %s", err.Error()))
	}

	signer, err := ssh.ParsePrivateKey(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse private key: %s", err.Error()))
	}
	return signer
}
