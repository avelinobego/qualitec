package util

import (
	"time"
)

type eventListener struct {
	stamp    time.Time
	receiver chan time.Time
	timeout  time.Duration
}

// EventsNotifier representa um notificador de eventos occorridos para listeners
// associados. Sempre que ocorrer um evento, todos os listeners serão notificados.
type EventsNotifier struct {
	// register é o channel responsável por receber as solicitações de registro/associação
	// de listeners.
	register chan *eventListener

	// unregister é o channel responsável por receber as solicitações de remoção de
	// listeners.
	unregister chan *eventListener

	// listeners armazena todos os listeners registrados/associados.
	listeners map[*eventListener]struct{}

	// notify recebe as notificações de eventos ocorridos.
	notify chan struct{}

	// stamp armazena o timestamp do último evento ocorrido. Sempre que um novo evento
	// for gerado, stamp conterá a data da ocorrência do evento.
	stamp time.Time

	// done quando fechado indica que a goroutine dos eventos
	// deverá ser finalizada
	done *DoneChannel
}

// NewEventsNotifier cria um novo notificador de eventos que enviará
// uma notificação cada vez que um novo evento for gerado.
//
// Sempre que um evento for gerado e o listener notificado, ele é automaticamente
// removido da lista de listeners e, para voltar a receber notificações, é
// preciso novamente se registrar como listener.
func NewEventsNotifier(done *DoneChannel) *EventsNotifier {
	e := EventsNotifier{
		register:   make(chan *eventListener),
		unregister: make(chan *eventListener),
		listeners:  make(map[*eventListener]struct{}),
		notify:     make(chan struct{}),
		stamp:      time.Now(),
		done:       done,
	}
	go e.run()
	return &e
}

// Notify notifica que houve um evento. A partir de um tempo configurável
// todos os listeners serão notificados desse evento.
// Caso evento tenha sido parado pelo DoneChannel, notificação é
// totalmente ignorada. Portanto, é seguro chamar mesmo quando eventos
// já tiverem sido parados.
func (e *EventsNotifier) Notify() {
	select {
	case e.notify <- struct{}{}:
	case <-e.done.Done():
	}
}

// WaitForEvent aguarda por uma notificação de um evento. Quando ocorrer um evento,
// retorna o timestamp do evento e o segundo retorno como true indicando que houve evento.
//
// O parâmetro stamp indica o timestamp da último evento que foi recebido.
// Caso seja a primeira vez de registro, passar time.Time{}.
//
// Já o parâmetro timeout indica o tempo que será aguardado até um evento ocorrer.
// Se não houver eventos nesse períoro, WaitForEvent retorna o timestamp = stamp e
// o segundo retorno como false indicando que não houve evento.
//
// Se último retorno for false, indica que o gerador de eventos não está mais ativo e que nenhum
// novo evento será gerado.
func (e *EventsNotifier) WaitForEvent(stamp time.Time, timeout time.Duration) (time.Time, bool, bool) {
	return e.registerAndWait(&eventListener{
		stamp:    stamp,
		receiver: make(chan time.Time, 1),
		timeout:  timeout,
	})
}

// WaitForNextEvent facilita a chamada WaitForEvent(time.Now(), timeout)
func (e *EventsNotifier) WaitForNextEvent(timeout time.Duration) (time.Time, bool, bool) {
	return e.WaitForEvent(time.Now(), timeout)
}

func (e *EventsNotifier) registerAndWait(l *eventListener) (i time.Time, ok bool, active bool) {
	select {
	case e.register <- l:
	case <-e.done.Done():
		return
	}

	if l.timeout > 0 {
		select {
		case i, ok = <-l.receiver:
			active = true
		case <-time.NewTimer(l.timeout).C:
			e.unregister <- l
			i = l.stamp
			active = true
		case <-e.done.Done():
		}
	} else {
		select {
		case i, ok = <-l.receiver:
			active = true
		case <-e.done.Done():
		}
	}
	return
}

func (e *EventsNotifier) run() {
	timer := time.NewTimer(0)
	timer.Stop()
	timerRunning := false

	for {
		select {
		case l := <-e.register:
			// Retorna imediatamente caso listener esteja desatualizado
			// e o timer não esteja rodando, fazendo com que ele não aguarde novas atualizações.
			if e.stamp.After(l.stamp) && !timerRunning {
				l.receiver <- e.stamp
			} else {
				e.listeners[l] = struct{}{}
			}

		case l := <-e.unregister:
			delete(e.listeners, l)

		case <-e.notify:
			// Novas notificações ativam a notificação dos listeners dentro de x período
			// de tempo. O objetivo é ter um limite de notificações para que elas
			// não sejam muito instantâneas e evitar notificações exageradas
			e.stamp = time.Now()
			if !timerRunning && len(e.listeners) > 0 {
				timer.Reset(500 * time.Millisecond)
				timerRunning = true
			}

		case <-timer.C:
			// Os listeners são notificados efetivamente aqui quando o timer for executado.
			for l := range e.listeners {
				if e.stamp.After(l.stamp) {
					l.receiver <- e.stamp
					delete(e.listeners, l)
				}
			}
			timer.Stop()
			timerRunning = false

		case <-e.done.Done():
			return
		}
	}
}
