package common

import (
	"bufio"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Channel struct {
	Name    string
	Clients map[*bufio.Writer]struct{}
	sync.RWMutex
}

type EventServer struct {
	Channels map[string]*Channel
	sync.RWMutex
}

func NewEventServer() *EventServer {
	return &EventServer{
		Channels: make(map[string]*Channel),
	}
}

func (s *EventServer) GetOrCreateChannel(name string) *Channel {
	s.RLock()
	channel, ok := s.Channels[name]
	s.RUnlock()
	if !ok {
		s.Lock()
		channel = &Channel{
			Name:    name,
			Clients: make(map[*bufio.Writer]struct{}),
		}
		s.Channels[name] = channel
		s.Unlock()
	}
	return channel
}

func (s *EventServer) RemoveClientFromChannel(channel *Channel, client *bufio.Writer) {
	channel.Lock()
	delete(channel.Clients, client)
	channel.Unlock()
}

func PrepareHeaderForSSE(c *fiber.Ctx) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
}

type EventMessage struct {
	ID     string    `json:"id"`
	Data   string    `json:"data"`
	Create time.Time `json:"create_date"`
}
