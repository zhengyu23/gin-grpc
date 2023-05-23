package models

import (
	"github.com/dgrijalva/jwt-go"
)

// CustomClaims 加密
type CustomClaims struct {
	ID          uint
	NickName    string
	AuthorityId uint
	jwt.StandardClaims
}
