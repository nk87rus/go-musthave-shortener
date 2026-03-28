package model

//generate:reset
type StorageRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OrigURL     string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

type Stats struct {
	URLs  int `json:"urls" db:"urls"`
	Users int `json:"users" db:"users"`
}
