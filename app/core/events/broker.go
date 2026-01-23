package events

import (
	"sync"
)

// Publisher defines an interface for publishing events.
type Publisher interface {
	Publish(event string)
}

// Subscriber defines an interface for subscribing to events.
type Subscriber interface {
	Subscribe() chan string
	Unsubscribe(clientChan chan string)
}

// Broker implements both Publisher and Subscriber for SSE.
type Broker struct {
	mu         sync.RWMutex
	clients    map[chan string]struct{}
	notifier   chan string
	register   chan chan string
	unregister chan chan string
}

// NewBroker creates a new Broker and starts its event loop.
func NewBroker() *Broker {
	b := &Broker{
		clients:    make(map[chan string]struct{}),
		notifier:   make(chan string, 1),
		register:   make(chan chan string),
		unregister: make(chan chan string),
	}
	go b.run()
	return b
}

// run is the main event loop for the broker.
func (b *Broker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = struct{}{}
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
			}
			b.mu.Unlock()
		case event := <-b.notifier:
			b.mu.RLock()
			for client := range b.clients {
				// Non-blocking send
				select {
				case client <- event:
				default:
				}
			}
			b.mu.RUnlock()
		}
	}
}

// Publish sends an event to all subscribed clients.
func (b *Broker) Publish(event string) {
	b.notifier <- event
}

// Subscribe creates a new client channel and registers it.
func (b *Broker) Subscribe() chan string {
	clientChan := make(chan string)
	b.register <- clientChan
	return clientChan
}

// Unsubscribe removes a client channel.
func (b *Broker) Unsubscribe(clientChan chan string) {
	b.unregister <- clientChan
}
