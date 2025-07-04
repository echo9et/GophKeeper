package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Keeper interface {
	Remove(key string) error
	UpdateRecord(key, value string) error
	Print()
}

func ReadCmd(k Keeper) {
	fmt.Println("Добро пожаловать в GophKeeper")
	fmt.Println("Введите команду (например: new KEY VALUE или del KEY или print):")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {

		case "new":
			if len(parts) != 3 {
				fmt.Println("Используйте: new KEY VALUE")
				continue
			}
			key := parts[1]
			value := parts[2]
			k.UpdateRecord(key, value)

		case "del":
			if len(parts) != 2 {
				fmt.Println("Используйте: del KEY")
			}
			key := parts[1]
			k.Remove(key)
			continue

		case "print":
			k.Print()
			continue

		default:
			fmt.Println("Неизвестная команда. Доступные команды: new, del, print")
		}
	}
}
