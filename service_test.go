package main

import (
	"errors"
	"sync"
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

type blockingProcessor struct {
	started chan struct{}
	release chan struct{}
}

func (b *blockingProcessor) Process(password string) (Result, error) {
	if err := ValidatePassword(password); err != nil {
		return Result{}, err
	}
	b.started <- struct{}{}
	<-b.release
	return Result{Length: len([]rune(password))}, nil
}

func TestConcurrentOverlapAndInvalidSubmit(t *testing.T) {
	proc := &blockingProcessor{
		started: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	svc := NewService(proc)
	svc.Start(2)
	t1, err := svc.Create("ValidPass123!")
	if err != nil {
		t.Fatalf("ошибка создания задачи: %v", err)
	}
	t2, err := svc.Create("AnotherValidPass456!")
	if err != nil {
		t.Fatalf("ошибка создания задачи: %v", err)
	}

	if err := svc.Submit(t1.ID); err != nil {
		t.Fatalf("ошибка отправки задачи 1: %v", err)
	}

	if err := svc.Submit(t1.ID); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("ожидалась ошибка ErrInvalidTransition при повторном Submit, получили: %v", err)
	}

	if err := svc.Submit(t2.ID); err != nil {
		t.Fatalf("ошибка отправки задачи 2: %v", err)
	}

	<-proc.started
	<-proc.started
	close(proc.release)

	for _, id := range []int{t1.ID, t2.ID} {
		for {
			got, err := svc.Get(id)
			if err != nil {
				t.Fatalf("ошибка получения задачи %d: %v", id, err)
			}
			if got.Status == StatusCompleted {
				if got.Result == nil {
					t.Errorf("задача %d: результат nil", id)
				}
				break
			}
		}
	}

	if err := svc.Submit(t1.ID); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("ожидалась ошибка ErrInvalidTransition после завершения задачи, получили: %v", err)
	}

	svc.Stop()

}

func TestConcurrent(t *testing.T) {
	svc := NewService(PasswordProcessor{})
	svc.Start(3)

	const total = 10
	var wg sync.WaitGroup
	wg.Add(total)

	for i := 0; i < total; i++ {
		go func(idx int) {
			defer wg.Done()

			task, err := svc.Create("ConcurrentPass123!")
			if err != nil {
				t.Errorf("ошибка создания задачи (итерация %d): %v", idx, err)
				return
			}
			if err := svc.Submit(task.ID); err != nil {
				t.Errorf("ошибка отправки задачи (итерация %d): %v", idx, err)
				return
			}
			if _, err := svc.Get(task.ID); err != nil {
				t.Errorf("ошибка получения задачи (итерация %d): %v", idx, err)
			}
		}(i)
	}

	wg.Wait()
	svc.Stop()
}
