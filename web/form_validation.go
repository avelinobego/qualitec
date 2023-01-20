package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type FormValidation struct {
	errors [][3]interface{}
	r      *http.Request
}

func NewFormValidation(r *http.Request) *FormValidation {
	return &FormValidation{
		errors: make([][3]interface{}, 0),
		r:      r,
	}
}

func (f *FormValidation) RequiredString(field string) string {
	v := f.r.FormValue(field)
	if v == "" {
		f.errors = append(f.errors, [3]interface{}{20, field, "Campo obrigatório!"})
	}
	return v
}

func (f *FormValidation) RequiredInt(field string) int {
	v := f.r.FormValue(field)
	if v == "" {
		f.errors = append(f.errors, [3]interface{}{20, field, "Campo obrigatório!"})
		return 0
	}
	vInt, err := strconv.ParseInt(v, 10, 0)
	if err != nil {
		f.errors = append(f.errors, [3]interface{}{20, field, err.Error()})
	}
	return int(vInt)
}

// RequiredUint converte uma string em uint. Armazena eventuais erros no
// objeto FormValidation para validação posterior.
//
// Caso a string seja vazia, o erro "20 - Campo obrigatório!" é armazenado
// e o número 0 é retornado
//
// Outros erros podem ser gerados pela função strconv.ParseUint e o retorno é dela tambén.
func (f *FormValidation) RequiredUint(field string, base int, bitSize int) uint64 {
	v := f.r.FormValue(field)
	if v == "" {
		f.errors = append(f.errors, [3]interface{}{20, field, "Campo obrigatório!"})
		return 0
	}
	vInt, err := strconv.ParseUint(v, base, bitSize)
	if err != nil {
		f.errors = append(f.errors, [3]interface{}{20, field, err.Error()})
	}
	return vInt
}

// RequiredDateTimeLayout valida a data/hora no formato especificado em layout e a retorna
func (f *FormValidation) RequiredDateTimeLayout(field string, notOlderThan time.Duration, layout string) *time.Time {
	var err error
	v := f.r.FormValue(field)
	if v == "" {
		f.errors = append(f.errors, [3]interface{}{20, field, "Campo obrigatório!"})
		return nil
	}
	d, err := time.Parse(layout, v)
	if err != nil {
		f.errors = append(f.errors, [3]interface{}{20, field, "Data/hora inválida: " + err.Error()})
		return nil
	}
	if notOlderThan > 0 && time.Now().After(d.Add(notOlderThan)) {
		f.errors = append(f.errors, [3]interface{}{20, field, "Data muito antiga"})
		return nil
	}
	return &d
}

// RequiredDateTime valida a data/hora no formato AAAA-DD-MM HH:MM e a retorna
func (f *FormValidation) RequiredDateTime(field string, notOlderThan time.Duration) (date *time.Time) {
	return f.RequiredDateTimeLayout(field, notOlderThan, "2006-01-02 15:04")
}

// RequiredDate valida a data no formato AAAA-DD-MM e a retorna
func (f *FormValidation) RequiredDate(field string, notOlderThan time.Duration) (date *time.Time) {
	return f.RequiredDateTimeLayout(field, notOlderThan, "2006-01-02")
}

func (f *FormValidation) RequiredTime(field string) (t time.Time) {
	var err error
	v := f.r.FormValue(field)
	if v == "" {
		f.errors = append(f.errors, [3]interface{}{20, field, "Campo obrigatório!"})
		return
	}
	t, err = time.Parse("15:04", v)
	if err != nil {
		f.errors = append(f.errors, [3]interface{}{20, field, "Hora inválida!"})
		return
	}
	return
}

func (f *FormValidation) Dispatch(w http.ResponseWriter) bool {
	hasError := len(f.errors) > 0
	if hasError {
		js, err := json.Marshal(f.errors)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return true
		}

		w.Header().Set("x-celus-error", "1")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(js))
		return true
	}
	return false
}
