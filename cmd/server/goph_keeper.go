package main

import (
	"fmt"

	server "GophKeeper.ru/internal/server/http"
	"GophKeeper.ru/internal/server/storage"
)

type Keeper struct {
	databse *storage.Database
	server  *server.Server
}

// NewKeeper(addrDatabase)  (*Keeper, error) конструктор сервера хранилища
func NewKeeper(addrDatabase string) (*Keeper, error) {
	db, err := storage.New(addrDatabase)
	if err != nil {
		return nil, fmt.Errorf("error connect database %s", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error ping database %s", err)
	}
	defer db.Stop()

	s, err := server.New(db)

	if err != nil {
		return nil, fmt.Errorf("error create server %s", err)
	}

	return &Keeper{
		databse: db,
		server:  s}, nil
}

// NewKeeper(addrDatabase)  (*Keeper, error) конструктор сервера хранилища
func (k *Keeper) Run(addr string) error {
	return k.server.Run(addr)
}

func (k *Keeper) Stop() error {
	k.server.Stop()

	return nil
}
