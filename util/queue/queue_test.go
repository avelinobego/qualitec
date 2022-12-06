package queue

import "testing"

type MockValue struct {
	id    string
	value int
}

func (m *MockValue) UniqueQueueID() interface{} {
	return m.id
}

func TestUniqueQueueOrder(t *testing.T) {
	q := NewUniqueQueue()
	q.Add(&MockValue{"c", 0}, ModeIgnoreIfExists)
	q.Add(&MockValue{"b", 0}, ModeIgnoreIfExists)
	q.Add(&MockValue{"a", 0}, ModeIgnoreIfExists)

	// Testa a ordem
	v, ok := q.Top().(*MockValue)
	if !ok {
		t.Errorf("Element not found")
	}
	if v.id != "c" {
		t.Errorf("Expected ID %v, found ID %v", "c", v.id)
	}

	v, ok = q.Top().(*MockValue)
	if !ok {
		t.Errorf("Element not found")
	}
	if v.id != "b" {
		t.Errorf("Expected ID %v, found ID %v", "b", v.id)
	}

	v, ok = q.Top().(*MockValue)
	if !ok {
		t.Errorf("Element not found")
	}
	if v.id != "a" {
		t.Errorf("Expected ID %v, found ID %v", "a", v.id)
	}

	v, ok = q.Top().(*MockValue)
	if ok {
		t.Errorf("Expected nothing, got ID %v", v.id)
	}
}
func TestUniqueQueueUpdateDuplicity(t *testing.T) {
	q := NewUniqueQueue()

	// Testa substituição de duplicidade
	q.Add(&MockValue{"c", 0}, ModeUpdateIfExists)
	q.Add(&MockValue{"b", 0}, ModeUpdateIfExists)
	q.Add(&MockValue{"a", 1}, ModeUpdateIfExists)
	q.Add(&MockValue{"a", 2}, ModeUpdateIfExists)
	q.Top()
	q.Top()

	v, ok := q.Top().(*MockValue)
	if !ok {
		t.Errorf("Element not found")
	}
	if v.id != "a" {
		t.Errorf("Expected ID %v, found ID %v", "a", v.id)
	}
	if v.value != 2 {
		t.Errorf("Expected value %v, found value %v", 2, v.value)
	}

	v, ok = q.Top().(*MockValue)
	if ok {
		t.Errorf("Expected nothing, got ID %v", v.id)
	}

}

func TestUniqueQueueDiscardDuplicity(t *testing.T) {
	q := NewUniqueQueue()
	// Testa o descarte de duplicidade
	q.Add(&MockValue{"c", 0}, ModeIgnoreIfExists)
	q.Add(&MockValue{"b", 0}, ModeIgnoreIfExists)
	q.Add(&MockValue{"a", 1}, ModeIgnoreIfExists)
	q.Add(&MockValue{"a", 2}, ModeIgnoreIfExists)
	q.Top()
	q.Top()

	v, ok := q.Top().(*MockValue)
	if !ok {
		t.Errorf("Element not found")
	}
	if v.id != "a" {
		t.Errorf("Expected ID %v, found ID %v", "a", v.id)
	}
	if v.value != 1 {
		t.Errorf("Expected value %v, found value %v", 1, v.value)
	}

	v, ok = q.Top().(*MockValue)
	if ok {
		t.Errorf("Expected nothing, got ID %v", v.id)
	}
}
