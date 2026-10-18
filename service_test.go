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
