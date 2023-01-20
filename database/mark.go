package database

type Marker struct {
	changes map[string]interface{}
}

func (m *Marker) Mark(key string, v interface{}) {
	if m.changes == nil {
		m.changes = make(map[string]interface{})
	}
	m.changes[key] = v
}

func (m *Marker) Marks() map[string]interface{} {
	return m.changes
}

func (m *Marker) UnMark() {
	m.changes = nil
}

func (m *Marker) MakeNamedSQL() (sql string) {
	if m.Marks() != nil {
		for k := range m.Marks() {
			sql += k + "=:" + k + ", "
		}
		sql = sql[0 : len(sql)-2]
	}
	return
}
