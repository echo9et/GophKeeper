package entities

import "context"

type AuthManager interface {
	CompareUser(hash string) error
	User(ctx context.Context, login string) (*User, error)
	UserFromID(ctx context.Context, id int) (*User, error)
}
