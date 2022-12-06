package web

import (
	"bytes"
	"html/template"
	"math"

	"celus-ti.com.br/qualitec/util"
)

type pagination struct {
	Active bool
	URL    string
	Page   int
}

func Pagination(t *template.Template, linesPerPage, totalLines, currentPage int, f func(int) string) (html template.HTML) {
	totalButtons := 5
	sideButtons := int(math.Floor(float64(totalButtons) / float64(2)))

	totalPages := int(math.Ceil(float64(totalLines) / float64(linesPerPage)))
	startPage := util.Max(currentPage-sideButtons-util.Max(currentPage+sideButtons-totalPages, 0), 1)

	lastPage := util.Min(currentPage+sideButtons+util.Abs(util.Min(currentPage-sideButtons-1, 0)), totalPages)

	paginationData := []*pagination{}
	for i := startPage; i <= lastPage; i++ {
		paginationData = append(paginationData, &pagination{
			Active: i == currentPage,
			URL:    f(i),
			Page:   i,
		})
	}

	data := map[string]interface{}{
		"FirstURL": f(1),
		"LastURL":  f(totalPages),
		"URLs":     paginationData,
	}
	writer := new(bytes.Buffer)
	err := t.ExecuteTemplate(writer, "pagination.html", data)
	if err != nil {
		return template.HTML(template.HTMLEscaper(err.Error()))
	}
	return template.HTML(writer.String())
}
