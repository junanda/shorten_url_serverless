package model

import "time"

type User struct {
	IdUser         string    `json:"id,omitempty"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	Email          string    `json:"email"`
	RegisterDate   time.Time `json:"registerDate"`
	UserSubsscribe string    `json:"subscribe,omitempty"`
	UpdateDate     time.Time `json:"updateDate"`
	MaxShortUrl    int       `json:"max_short_month"`
	Role           string    `json:"role,omitempty"`
}

type UpdateUserRequest struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email"`
}

type ListMembersRequest struct {
	Role      string `json:"role"`
	Subscribe string `json:"subscribe"`
}

type ListMembersResponse struct {
	Data       []DataUser `json:"data"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"totalPages"`
}

type DataUser struct {
	Id             string `json:"id,omitempty"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	UserSubsscribe string `json:"subscribe,omitempty"`
	MaxShortUrl    int    `json:"max_short_month"`
}

type RecoveryRequest struct {
	Email string `json:"email"`
}
