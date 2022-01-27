package models

import (
	"time"
)

type History struct {
	URL          URL       `json:"url"`
	StatusCode   int       `json:"status"`
	Requested_at time.Time `json:"requested_at"`
}
