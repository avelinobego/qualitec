package queue

import (
	"sync"
)

const (
	// ModeIgnoreIfExists informa à fila que se o elemento já existir,
	// ignora a adição. Se não existir, adiciona no fim da fila
	ModeIgnoreIfExists = 0
	// ModeUpdateIfExists informa à fila que se o elemento já existir,
	// substitui o elemento existente. Se não existir, adiciona no fim da fila.
	ModeUpdateIfExists = 1
)

// UniqueQueueValue interface para os elementos da UniqueQueue (fila única).
// Basicamente é preciso retornar um valor único para o tipo de elemento
// que se deseja colocar na fila.
type UniqueQueueValue interface {
	// Retorna um valor único para o elemento da fila
	UniqueQueueID() interface{}
}

// UniqueQueue implementa o conceito de fila única: elementos com o mesmo ID
// são substituídos pelos mais recentemente adicionados.
// A implementação pode ser usada em concorrência.
type UniqueQueue struct {
	queue []UniqueQueueValue
	index map[interface{}]int
	mutex *sync.RWMutex
}

// NewUniqueQueue inicializa uma fila única
func NewUniqueQueue() *UniqueQueue {
	return &UniqueQueue{
		queue: make([]UniqueQueueValue, 0),
		index: make(map[interface{}]int),
		mutex: &sync.RWMutex{},
	}
}

// Add adiciona um elemento na fila de acordo com o mode.
func (q *UniqueQueue) Add(value UniqueQueueValue, mode int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	index, ok := q.index[value.UniqueQueueID()]
	if ok {
		if mode == ModeUpdateIfExists {
			q.queue[index] = value
		}
	} else {
		q.queue = append(q.queue, value)
		q.index[value.UniqueQueueID()] = len(q.queue) - 1
	}
}

// Top remove um elemento do começo da fila
func (q *UniqueQueue) Top() (u UniqueQueueValue) {
	q.mutex.RLock()
	if len(q.queue) > 0 {
		u = q.queue[0]
		// Remove o valor e o índice das estruturas internas
		q.queue = q.queue[1:]
		delete(q.index, u.UniqueQueueID())
	}
	q.mutex.RUnlock()
	return
}
