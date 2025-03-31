package models

type Creds struct {
	Username string `json:"username" validate:"required min=3 max=255 utf8"`
	Password string `json:"password" validate:"required min=3 max=32"`
}
