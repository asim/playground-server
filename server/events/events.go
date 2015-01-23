package events

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	Error   = "error"
	Message = "message"
	Status  = "status"
)

type Event struct {
	Id        string
	Body      string
	Type      string
	Timestamp int64
}

type Streamer struct {
	pool *redis.Pool

	sync.RWMutex
	subscribers map[<-chan Event]redis.PubSubConn
}

var (
	streamer = newStreamer()
)

func newStreamer() *Streamer {
	host := os.Getenv("PLAYGROUND_REDIS_SERVICE_HOST")
	port := os.Getenv("PLAYGROUND_REDIS_SERVICE_PORT")

	if len(host) == 0 {
		host = "127.0.0.1"
	}

	if len(port) == 0 {
		port = "6379"
	}

	pool := redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", host+":"+port)
	}, 5)

	return &Streamer{
		pool:        pool,
		subscribers: make(map[<-chan Event]redis.PubSubConn),
	}
}

func subscriber(psc redis.PubSubConn, ch chan<- Event) {
	for {
		switch n := psc.Receive().(type) {
		case redis.Message:
			var ev Event
			err := json.Unmarshal(n.Data, &ev)
			if err != nil {
				continue
			}
			ch <- ev
		case redis.PMessage:
			var ev Event
			err := json.Unmarshal(n.Data, &ev)
			if err != nil {
				continue
			}
			ch <- ev
		case redis.Subscription:
			if n.Count == 0 {
				close(ch)
				return
			}
		case error:
			return
		}
	}
}

func (s *Streamer) Subscribe(topic string) <-chan Event {
	ch := make(chan Event, 1)
	c := s.pool.Get()
	psc := redis.PubSubConn{Conn: c}
	psc.PSubscribe("topics:" + topic)

	s.Lock()
	s.subscribers[ch] = psc
	s.Unlock()

	go subscriber(psc, ch)

	return ch
}

func (s *Streamer) Unsubscribe(topic string, ch <-chan Event) {
	s.Lock()
	defer s.Unlock()

	sub, ok := s.subscribers[ch]
	if !ok {
		return
	}

	sub.PUnsubscribe()
	sub.Close()
	delete(s.subscribers, ch)
}

func (s *Streamer) Send(topic string, event Event) {
	if len(event.Id) == 0 {
		event.Id = topic
	}

	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	c := s.pool.Get()
	defer c.Close()

	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	c.Do("PUBLISH", "topics:"+topic, b)
}

func Receive(topic string, stream io.Reader) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		text := scanner.Text()
		streamer.Send(topic, Event{Id: topic, Body: text, Type: Message})
	}
	if err := scanner.Err(); err != nil {
		return
	}
}

func Send(topic string, event Event) {
	streamer.Send(topic, event)
}

func Subscribe(topic string) <-chan Event {
	return streamer.Subscribe(topic)
}

func Unsubscribe(topic string, ch <-chan Event) {
	streamer.Unsubscribe(topic, ch)
}
