package entities

import "context"

type Record struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type Update struct {
	Value int      `json:"value"`
	Data  []Record `json:"data"`
}

func NewUpdate() *Update {
	return &Update{}
}

type DataManager interface {
	GetData(ctx context.Context, id int) (*Update, error)
	GetCountUpdate(ctx context.Context, userID int) (int, error)
	UpdateRecord(ctx context.Context, userID int, r Record) error
	RemoveRecord(ctx context.Context, userID int, key string) error
}
