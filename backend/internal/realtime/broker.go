package realtime

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// Broker fans out lightweight notifications to SSE subscribers per project.
type Broker struct {
	mu   sync.Mutex
	subs map[uuid.UUID][]chan struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: make(map[uuid.UUID][]chan struct{})}
}

// Subscribe returns a channel that receives a signal when tasks in projectID change.
// Unsubscribe happens when ctx is done.
func (b *Broker) Subscribe(ctx context.Context, projectID uuid.UUID) <-chan struct{} {
	ch := make(chan struct{}, 8)
	b.mu.Lock()
	b.subs[projectID] = append(b.subs[projectID], ch)
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		list := b.subs[projectID]
		for i, c := range list {
			if c == ch {
				b.subs[projectID] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(b.subs[projectID]) == 0 {
			delete(b.subs, projectID)
		}
		b.mu.Unlock()
	}()

	return ch
}

// PublishTasksChanged notifies all subscribers for a project (non-blocking per subscriber).
func (b *Broker) PublishTasksChanged(projectID uuid.UUID) {
	b.mu.Lock()
	list := append([]chan struct{}(nil), b.subs[projectID]...)
	b.mu.Unlock()
	for _, ch := range list {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
