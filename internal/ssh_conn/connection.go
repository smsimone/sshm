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
	Label   string  `json:"label"`
	Host    string  `json:"host"`
	Port    int     `json:"port"`
	Profile *string `json:"profile"`
}

func (c *Connection) String(asJson bool) string {
	basic := fmt.Sprintf("(%s) %s@%s:%d", c.Label, c.GetProfile().Username, c.Host, c.Port)
	if !asJson {
		return basic
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return basic
	}
	return string(bytes)
}

func (c *Connection) GetProfile() Profile {
	profile, err := GetProfile(*c.Profile)
	if err != nil {
		panic(err)
	}
	return *profile
}

func (c *Connection) AuthMethods() []ssh.AuthMethod {
	methods := []ssh.AuthMethod{}

	profile := c.GetProfile()

	if profile.Password != nil {
		methods = append(methods, ssh.Password(*profile.Password))
	}

	if profile.PvtKey != nil {
		methods = append(methods, ssh.PublicKeys(profile.PvtKey.signer()))
	}

	return methods
}

func (pvt *PrivateKey) signer() ssh.Signer {
	content, err := base64.StdEncoding.DecodeString(pvt.Content)
	if err != nil {
		panic(fmt.Sprintf("Failed to read private key content: %s", err.Error()))
	}

	if pvt.Passphrase != nil && len(*pvt.Passphrase) > 0 {
		signer, err := ssh.ParsePrivateKeyWithPassphrase(content, []byte(*pvt.Passphrase))
		if err != nil {
			panic(fmt.Sprintf("Failed to parse private key: %s", err.Error()))
		}
		return signer
	}

	signer, err := ssh.ParsePrivateKey(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse private key: %s", err.Error()))
	}
	return signer
}
