package sshconn

type Profile struct {
	Id       *int        `json:"id"`
	Label    string      `json:"label"`
	Username string      `json:"username"`
	PvtKey   *PrivateKey `json:"private_key"`
	Password *string     `json:"password"`
}
