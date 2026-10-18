package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	svc := NewService(PasswordProcessor{})
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Введи пароль (или q для выхода): ")
		line, _ := reader.ReadString('\n')
		password := strings.TrimRight(line, "\r\n")

		if password == "q" {
			break
		}

		task, err := svc.Create(password)
		if err != nil {
			fmt.Println("ошибка:", err)
			continue
		}

		if err := svc.Process(task.ID); err != nil {
			fmt.Println("ошибка:", err)
			continue
		}

		task, _ = svc.Get(task.ID)
		r := task.Result

		fmt.Printf("задача %d: %s\n", task.ID, task.Status)
		fmt.Printf("  длина: %d\n", r.Length)
		fmt.Printf("  есть цифра: %v\n", r.HasDigit)
		fmt.Printf("  есть заглавная: %v\n", r.HasUpper)
		fmt.Printf("  есть символ: %v\n", r.HasSymbol)
	}
}
