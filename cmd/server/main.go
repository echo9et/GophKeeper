package main

import (
	"fmt"
)

func main() {
	config, err := GetConfig()
	if err != nil {
		panic(err)
	}

	keeper, err := NewKeeper(config.AddrDatabase)
	if err != nil {
		panic(err)
	}

	if err := keeper.Run(config.AddrServer); err != nil {
		panic(err)
	}

	fmt.Println("down server")
}
