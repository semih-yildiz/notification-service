package create

// Command for creating a single notification.
type Command struct {
	Recipient      string
	Channel        string
	Content        string
	Priority       string
	IdempotencyKey *string
}

// BatchItem for one notification in a batch.
type BatchItem struct {
	Recipient string
	Channel   string
	Content   string
	Priority  string
}

// BatchCommand for creating a batch of notifications (max 1000).
type BatchCommand struct {
	Items          []BatchItem
	IdempotencyKey *string
}
