package main

import (
	"errors"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	svc := NewService(PasswordProcessor{})

	task, err := svc.Create("Hello123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != StatusPending {
		t.Errorf("want status pending, got %s", task.Status)
	}

	got, err := svc.Get(task.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Password != "Hello123!" {
		t.Errorf("want password Hello123!, got %s", got.Password)
	}
}

func TestCreateInvalid(t *testing.T) {
	svc := NewService(PasswordProcessor{})

	_, err := svc.Create("abc")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("want ErrInvalidInput, got %v", err)
	}
}

func TestGetNotFound(t *testing.T) {
	svc := NewService(PasswordProcessor{})

	_, err := svc.Get(999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestProcess(t *testing.T) {
	svc := NewService(PasswordProcessor{})
	task, _ := svc.Create("Hello123!")

	if err := svc.Process(task.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.Get(task.ID)
	if got.Status != StatusCompleted {
		t.Errorf("want status completed, got %s", got.Status)
	}
	if got.Result == nil {
		t.Fatal("result is nil")
	}
	if got.Result.Length != 9 {
		t.Errorf("want length 9, got %d", got.Result.Length)
	}
	if !got.Result.HasDigit {
		t.Error("want HasDigit true")
	}
}

func TestProcessTwice(t *testing.T) {
	svc := NewService(PasswordProcessor{})
	task, _ := svc.Create("Hello123!")

	_ = svc.Process(task.ID)
	err := svc.Process(task.ID)

	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("want ErrInvalidTransition, got %v", err)
	}
}

func TestProcessNotFound(t *testing.T) {
	svc := NewService(PasswordProcessor{})

	err := svc.Process(999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestMultipleTasks(t *testing.T) {
	svc := NewService(PasswordProcessor{})

	t1, _ := svc.Create("Hello123!")
	t2, _ := svc.Create("Abcdefgh!")

	if t1.ID == t2.ID {
		t.Fatal("tasks must have different IDs")
	}

	_ = svc.Process(t1.ID)

	got1, _ := svc.Get(t1.ID)
	got2, _ := svc.Get(t2.ID)

	if got1.Status != StatusCompleted {
		t.Errorf("task 1: want completed, got %s", got1.Status)
	}
	if got2.Status != StatusPending {
		t.Errorf("task 2: want pending, got %s", got2.Status)
	}
	if got1.Result.HasDigit != true {
		t.Error("task 1: want HasDigit true")
	}
	if got2.Result != nil {
		t.Error("task 2: result must stay nil")
	}
}

func TestQueueFull(t *testing.T) {
	svc := NewService(PasswordProcessor{})

	for i := 0; i < 10; i++ {
		task, _ := svc.Create("ValidPass123!")
		if err := svc.Submit(task.ID); err != nil {
			t.Fatalf("неожиданная ошибка при заполнении очереди: %v", err)
		}
	}

	task11, _ := svc.Create("ValidPass123!")
	err := svc.Submit(task11.ID)

	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("ожидалась ошибка ErrQueueFull, получили: %v", err)
	}
}

func TestSubmitAfterStop(t *testing.T) {
	svc := NewService(PasswordProcessor{})

	svc.Stop()

	task, _ := svc.Create("ValidPass123!")
	err := svc.Submit(task.ID)

	if !errors.Is(err, ErrStopped) {
		t.Errorf("ожидалась ошибка ErrStopped, получили: %v", err)
	}
}

func TestConcurrentProcessing(t *testing.T) {
	svc := NewService(PasswordProcessor{})
	
	svc.Start(3)

	var ids []int
	for i := 0; i < 5; i++ {
		task, err := svc.Create("GoRules2026!") 
		if err != nil {
			t.Fatalf("ошибка создания: %v", err)
		}
		if err := svc.Submit(task.ID); err != nil {
			t.Fatalf("ошибка отправки в очередь: %v", err)
		}
		ids = append(ids, task.ID)
	}

	svc.Stop()

	for _, id := range ids {
		got, err := svc.Get(id)
		if err != nil {
			t.Fatalf("ошибка получения задачи %d: %v", id, err)
		}
		if got.Status != StatusCompleted {
			t.Errorf("задача %d: ожидался статус completed, получили %s", id, got.Status)
		}
		if got.Result == nil {
			t.Errorf("задача %d: результат nil", id)
		}
	}
}
