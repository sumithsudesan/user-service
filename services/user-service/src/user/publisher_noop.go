package user

// Used in test -noopPublisher discards all events.
type noopPublisher struct{}

// NewNoopPublisher returns a Publisher that discards all events.
func NewNoopPublisher() Publisher {
	return &noopPublisher{}
}

// Publish discards the event and returns nil.
func (p *noopPublisher) Publish(_ Event) error {
	return nil
}
