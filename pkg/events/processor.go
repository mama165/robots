package events

// Processor
// Each kind of event has his own processor
// Based on the Chain of responsibility pattern
type Processor interface {
	Handle(event Event)
}
