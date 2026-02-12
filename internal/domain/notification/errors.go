package notification

import "errors"

var (
	ErrNotFound         = errors.New("notification or batch not found")
	ErrInvalidChannel   = errors.New("invalid channel")
	ErrInvalidPriority  = errors.New("invalid priority")
	ErrInvalidContent   = errors.New("invalid content: character limits or required fields")
	ErrDuplicateRequest = errors.New("duplicate request: idempotency key already used")
	ErrBatchTooLarge    = errors.New("batch size exceeds maximum (1000)")
	ErrAlreadyTerminal  = errors.New("notification already in terminal state")
)
