package database

import (
	"strings"
)

// Gera bind vars para cláusula IN do SQL. Por exemplo, MakeIn(3) == "(?,?,?)"
func MakeIn(count uint) string {
	// Casos comuns são tradados individualmente por questões de desempenho
	switch count {
	case 1:
		return "(?)"
	case 2:
		return "(?,?)"
	case 3:
		return "(?,?,?)"
	case 4:
		return "(?,?,?,?)"
	case 5:
		return "(?,?,?,?,?)"
	case 0:
		return "()"
	default:
		return "(?" + strings.Repeat(",?", int(count-1)) + ")"
	}
}

type Clauses []string

func (clauses Clauses) MakeAndWhere(putWhere bool) (sqlWhere string) {
	for _, w := range clauses {
		sqlWhere += w + " AND "
	}
	if len(sqlWhere) > 0 {
		sqlWhere = sqlWhere[:len(sqlWhere)-5]
		if putWhere {
			sqlWhere = "WHERE " + sqlWhere
		}
	}
	return
}
