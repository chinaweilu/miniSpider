package queue

import (
	"container/list"
	"time"
)

const MAX_SERVER_MAP_NUM int = 100

type Queue struct {
	list *list.List
}

type ServerMap struct {
	host       string
	lastOpTime time.Time
	interval   time.Duration
	timeout    time.Duration
	count      int
}

func NewQueue() *Queue {
	list := list.New()
	return &Queue{list: list}
}

func (queue *Queue) Push(value interface{}) {
	queue.list.PushBack(value)
}

func (queue *Queue) Pop() interface{} {
	e := queue.list.Front()
	if e != nil {
		queue.list.Remove(e)
		return e.Value
	}
	return nil
}

func (queue *Queue) Peak() interface{} {
	e := queue.list.Front()
	if e != nil {
		return e.Value
	}

	return nil
}

func (queue *Queue) Len() int {
	return queue.list.Len()
}

func (queue *Queue) Empty() bool {
	return queue.list.Len() == 0
}

func NewServerMap(host string, interval, timeout time.Duration) *ServerMap {
	s := &ServerMap{
		host:     host,
		interval: interval,
		timeout:  timeout,
	}

	return s

}

func (s *ServerMap) Host() string {
	return s.host
}

func (s *ServerMap) LastOpTime() time.Time {
	return s.lastOpTime
}

func (s *ServerMap) SetLastOpTime(time time.Time) time.Time {
	s.lastOpTime = time
	return s.lastOpTime
}

func (s *ServerMap) Interval() time.Duration {
	return s.interval
}

func (s *ServerMap) Timeout() time.Duration {
	return s.timeout
}

func (s *ServerMap) AddCount() int {
	s.count++
	return s.count
}

func (s *ServerMap) Count() int {
	return s.count
}

func (s *ServerMap) Sleep() {
	if time.Since(s.lastOpTime) < s.interval && s.count > MAX_SERVER_MAP_NUM {
		time.Sleep(s.interval - time.Since(s.lastOpTime))
	}
}
