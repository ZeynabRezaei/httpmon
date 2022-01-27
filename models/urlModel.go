package models

type URL struct {
	URL       string `json:"url" validate:"required""`
	Threshold int    `json:"threshold" validate:"required,min=1,max=2""`
	Failed    int    `json:"failed"`
	Succeed   int    `json:"succeed"`
}
