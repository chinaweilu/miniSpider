package queue

import "testing"

func TestQueue(t *testing.T) {
	queue := NewQueue()
	queue.Push(1)
	queue.Push(2)
	queue.Push(3)
	queue.Push(4)

	len := queue.Len()
	if len != 4 {
		t.Errorf("queue.Len() failed. Got %d, expected 4.", len)
	}

	value := queue.Peak().(int)
	if value != 1 {
		t.Errorf("queue.Peak() failed. Got %d, expected 4.", value)
	}

	value = queue.Pop().(int)
	if value != 1 {
		t.Errorf("queue.Pop() failed. Got %d, expected 4.", value)
	}

	len = queue.Len()
	if len != 3 {
		t.Errorf("queue.Len() failed. Got %d, expected 3.", len)
	}

	value = stack.Peak().(int)
	if value != 2 {
		t.Errorf("stack.Peak() failed. Got %d, expected 3.", value)
	}

	value = queue.Pop().(int)
	if value != 2 {
		t.Errorf("queue.Pop() failed. Got %d, expected 3.", value)
	}

	value = queue.Pop().(int)
	if value != 3 {
		t.Errorf("queue.Pop() failed. Got %d, expected 2.", value)
	}

	empty := queue.Empty()
	if empty {
		t.Errorf("queue.Empty() failed. Got %d, expected false.", empty)
	}

	value = queue.Pop().(int)
	if value != 4 {
		t.Errorf("queue.Pop() failed. Got %d, expected 1.", value)
	}

	empty = queue.Empty()
	if !empty {
		t.Errorf("queue.Empty() failed. Got %d, expected true.", empty)
	}

	nilValue := queue.Peak()
	if nilValue != nil {
		t.Errorf("queue.Peak() failed. Got %d, expected nil.", nilValue)
	}

	nilValue = queue.Pop()
	if nilValue != nil {
		t.Errorf("queue.Pop() failed. Got %d, expected nil.", nilValue)
	}

	len = queue.Len()
	if len != 0 {
		t.Errorf("queue.Len() failed. Got %d, expected 0.", len)
	}
}
