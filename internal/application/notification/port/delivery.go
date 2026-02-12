package port

import "context"

type DeliveryRequest struct {
	To      string
	Channel string
	Content string
}

type DeliveryResponse struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type DeliveryClient interface {
	Deliver(ctx context.Context, req *DeliveryRequest) (*DeliveryResponse, int, error)
}
