package models

type AuthResult struct {
	Sub               string
	PreferredUsername string
	Err               error
}

type UserInfo struct {
	ID           string `json:"id"`
	Username     string `json:"name"`
	PlatformRole string `json:"platformRole"`
}
