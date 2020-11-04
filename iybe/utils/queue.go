package utils

import (
	"sync"
	"io"
	"os"
	"encoding/json"
	"fmt"
	"reflect"
)

type item interface{}

type Queue struct {
	items []item
	lock  sync.Mutex
}

func NewQ() *Queue {
	return &Queue{items: make([]item, 0)}
}

func (q *Queue) Add(items ...item) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items = append(q.items, items...)

	return
}

func (q *Queue) Remove() item {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	item := q.items[0]
	q.items = q.items[1:]

	return item
}

func (q *Queue) Size() int {
	return len(q.items)
}

func (q *Queue) RemoveUpTo(num int) []item {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	items := make([]item, num)
	i := 0
	for i < num && i < len(q.items) {
		items[i] = q.items[i]
		i++
	}
	q.items = q.items[i:]
	return items[:i]
}

func (q *Queue) PrintQueue() {
	fmt.Println("printing queue contents")
	for _, item := range q.items {
		fmt.Println(item)
	}
}

func (q *Queue) Contains(data interface{}) bool {
	for _, v := range q.items {
		if reflect.DeepEqual(data, v) {
			return true
		}
	}

	return false
}

func (q *Queue) Serialize(filename string) error {
	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	for _, id := range q.items {
		buf, err := json.Marshal(id)
		if err != nil {
			return err
		}
		_, err = f.Write(buf)
		if err != nil {
			return err
		}
	}

	return f.Close()
}

func Deserialize(filename string, sep int) (*Queue, error) {
	var id item
	q := NewQ()
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	buf := make([]byte, sep)
	for {
		_, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				return q, nil
			}
      return nil, err
		}
		json.Unmarshal(buf, &id)
		q.Add(id)
	}

}
