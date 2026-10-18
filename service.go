package main

import "fmt"

type Service struct {
	tasks     map[int]*Task
	nextID    int
	processor Processor
}

func NewService(p Processor) *Service {
	return &Service{
		tasks:     make(map[int]*Task),
		nextID:    1,
		processor: p,
	}
}

func (s *Service) Create(password string) (*Task, error) {
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	t := &Task{
		ID:       s.nextID,
		Password: password,
		Status:   StatusPending,
	}
	s.tasks[t.ID] = t
	s.nextID++

	return t, nil
}

func (s *Service) Get(id int) (*Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("%w: id %d", ErrNotFound, id)
	}
	return t, nil
}

func (s *Service) Process(id int) error {
	t, err := s.Get(id)
	if err != nil {
		return err
	}

	if t.Status != StatusPending {
		return fmt.Errorf("%w: cannot process task in status %q", ErrInvalidTransition, t.Status)
	}

	t.Status = StatusProcessing

	result, err := s.processor.Process(t.Password)
	if err != nil {
		t.Status = StatusPending
		return err
	}

	t.Result = &result
	t.Status = StatusCompleted

	return nil
}
