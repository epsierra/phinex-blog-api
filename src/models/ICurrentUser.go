package models

import "github.com/golang-jwt/jwt/v4"

// ICurrentUser represents the current user data accessible in routes
type ICurrentUser struct {
	UserId          string     `json:"userId,omitempty"`
	Email           string     `json:"email,omitempty"`
	FullName        string     `json:"fullName,omitempty"`
	Roles           []string   `json:"roles,omitempty"`
	IsAuthenticated bool       `json:"isAuthenticated,omitempty"`
	IP              string     `json:"ip,omitempty"`
	Status          UserStatus `json:"status,omitempty"`
	jwt.RegisteredClaims
}
