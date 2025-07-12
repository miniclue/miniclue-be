package dto

// PubSubPushRequest is the request body for a Pub/Sub push notification.
type PubSubPushRequest struct {
	Message      PubSubMessage `json:"message"`
	Subscription string        `json:"subscription"`
}

// PubSubMessage is the actual message from Pub/Sub.
type PubSubMessage struct {
	Data       string            `json:"data"` // Base64-encoded
	MessageID  string            `json:"messageId"`
	Attributes map[string]string `json:"attributes"`
}
