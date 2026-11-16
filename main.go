package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	svc := NewService(PasswordProcessor{})
	
	// 1. Нанимаем и запускаем 2 рабочих (воркеров), как просит препод
	svc.Start(2)
	
	// 2. Гарантируем, что перед выходом из программы мы дождемся завершения всех задач
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

		// Создаем задачу
		task, err := svc.Create(password)
		if err != nil {
			fmt.Println("ошибка:", err)
			continue
		}

		// 3. Кидаем задачу на ленту (в очередь) и мгновенно идем дальше
		if err := svc.Submit(task.ID); err != nil {
			fmt.Println("ошибка очереди:", err)
			continue
		}

		// Вытаскиваем задачу прямо сейчас, чтобы посмотреть статус.
		// Скорее всего она будет "queued" или "processing", так как воркер еще работает в фоне!
		task, _ = svc.Get(task.ID)
		fmt.Printf("-> Задача %d отправлена на конвейер. Текущий статус: %s\n", task.ID, task.Status)
	}
}