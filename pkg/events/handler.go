package events

// EventHandler
// Each kind of event has his own handler
// Based on the Chain of responsibility pattern
type EventHandler interface {
	Handle(event Event)
}
