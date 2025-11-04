package model

type StorageRecord struct {
	UUID     string `json:"uuid"`
	ShortURL string `json:"short_url"`
	OrigURL  string `json:"original_url"`
}
