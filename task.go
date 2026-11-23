package main

type Status string

const (
	StatusPending    Status = "pending"
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
)

type Result struct {
	Length    int
	HasDigit  bool
	HasUpper  bool
	HasSymbol bool
}

type Task struct {
	ID       int
	Password string
	Status   Status
	Result   *Result
}
