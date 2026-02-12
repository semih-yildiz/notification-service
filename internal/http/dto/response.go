package dto

// ListResponse for GET /notifications (paginated).
type ListResponse struct {
	Notifications interface{} `json:"notifications"`
	Total         int         `json:"total"`
}

// BatchWithNotificationsResponse for GET /batches/:id/notifications.
type BatchWithNotificationsResponse struct {
	Batch         interface{} `json:"batch"`
	Notifications interface{} `json:"notifications"`
}

// CancelBatchResponse for POST /batches/:id/cancel.
type CancelBatchResponse struct {
	Cancelled int `json:"cancelled"`
}
