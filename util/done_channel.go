package util

// DoneChannel é usado para cancelar tarefas em todos os nívels de goroutines,
// similar à classe context.Context, porém mais simples e direta.
type DoneChannel struct {
	done chan struct{}
}

// NewDoneChannel instancia uma nova estrutura
func NewDoneChannel() *DoneChannel {
	return &DoneChannel{done: make(chan struct{})}
}

// Close fecha o canal. Não pode ser chamada duas vezes
func (d *DoneChannel) Close() {
	close(d.done)
}

// Done retorna o canal usado para detectar se está fechado ou não.
// É ideal para ser usado em select
func (d *DoneChannel) Done() <-chan struct{} {
	return d.done
}
