package model

import "time"

type Analytics struct {
	AnalyticId string    `json:"id,omitempty"`
	ShortUrlId string    `json:"idshort"`
	Browser    string    `json:"browser"`
	IpRequest  string    `json:"iprequest"`
	AccessDate time.Time `json:"accessdate"`
}
