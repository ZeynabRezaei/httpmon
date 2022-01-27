package models

type URL struct {
	URL       string `json:"url"`
	Threshold int    `json:"threshold"`
	Failed    int    `json:"failed"`
	Succeed   int    `json:"succeed"`
}
