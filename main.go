package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	svc := NewService(PasswordProcessor{})

	svc.Start(2)

	defer svc.Stop()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("CP2: Воркер-пул запущен (2 горутины).")

	for {
		fmt.Print("Введи пароль (или q для выхода): ")
		line, _ := reader.ReadString('\n')
		password := strings.TrimRight(line, "\r\n")

		if password == "q" {
			fmt.Println("\nОстановка сервиса. Ждем завершения фоновых задач...")
			break
		}

		task, err := svc.Create(password)
		if err != nil {
			fmt.Println("ошибка:", err)
			continue
		}

		if err := svc.Submit(task.ID); err != nil {
			fmt.Println("ошибка очереди:", err)
			continue
		}

		task, _ = svc.Get(task.ID)
		fmt.Printf("-> Задача %d отправлена на конвейер. Текущий статус: %s\n", task.ID, task.Status)
	}
}
