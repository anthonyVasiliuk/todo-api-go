package jwt

import "os"

func GetJWTSecret() string {
	if os.Getenv("APP_ENV") != "production" {
		return os.Getenv("JWT_SECRET_TESTING")
	} else {
		return os.Getenv("JWT_SECRET")
	}
}
