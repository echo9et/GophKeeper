package main

import (
	"GophKeeper.ru/internal/client/cli"
)

func main() {
	config, err := GetConfig()
	if err != nil {
		panic(err)
	}

	gophKeeper, err := NewGophKeeper(config.Username, config.Password, config.AddrServer, config.SecretKey, config.CryptoKey)
	if err != nil {
		panic(err)
	}

	if err = gophKeeper.auth(); err != nil {
		panic("Ошибка авторизации:" + err.Error())
	}

	go cli.ReadCmd(gophKeeper)
	gophKeeper.syncData()
}
