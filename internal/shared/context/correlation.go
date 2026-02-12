package context

type correlationIDKey struct{}

func CorrelationIDKey() interface{} {
	return correlationIDKey{}
}
