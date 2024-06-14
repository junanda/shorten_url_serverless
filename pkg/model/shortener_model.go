package model

import "time"

type ShortenURLRequest struct {
	UserID string `json:"userid,omitempty"`
	URL    string `json:"url"`
}

type ShortURL struct {
	IDShort    string    `json:"id,omitempty"`
	UserID     string    `json:"iduser"`
	URL        string    `json:"url"`
	ShortURL   string    `json:"short_url"`
	CreateDate time.Time `json:"createdate"`
}
