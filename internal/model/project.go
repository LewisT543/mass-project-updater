package model

type Project struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	SSHURL string `json:"ssh_url_to_repo"`
}
