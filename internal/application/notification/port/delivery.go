package port

import "context"

type DeliveryRequest struct {
	To      string `json:"to"`
	Channel string `json:"channel"`
	Content string `json:"content"`
}

type DeliveryResponse struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type DeliveryClient interface {
	Deliver(ctx context.Context, req *DeliveryRequest) (*DeliveryResponse, int, error)
}
