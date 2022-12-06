package web

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	model "celus-ti.com.br/qualitec/database/model"
	database "celus-ti.com.br/qualitec/server/database/model"
)

type URLQuery struct {
	values        url.Values
	path          string
	clientsString []string
	Customers     []model.CustomerID
	Clients       []int16
	Sites         []database.SiteID
}

func NewURLQuery(values url.Values, path string) *URLQuery {
	values.Del("_") // Remove the jQuery value

	urlQuery := URLQuery{
		values: values,
		path:   path,
	}

	sitesString := values["s"]
	for _, v := range sitesString {
		n, _ := strconv.ParseInt(v, 10, 16)
		urlQuery.Sites = append(urlQuery.Sites, database.SiteID(n))
	}

	clientsString := values["c"]
	for _, v := range clientsString {
		n, _ := strconv.ParseInt(v, 10, 16)
		urlQuery.Clients = append(urlQuery.Clients, int16(n))
	}

	urlQuery.clientsString = clientsString
	if len(urlQuery.clientsString) > 0 {
		urlQuery.Customers = model.CustomerSliceStrToCustomerID(urlQuery.clientsString, true)
	}

	return &urlQuery
}
func (u *URLQuery) Encode() string {
	return encode(u.values, u.path)
}

func (u *URLQuery) GenURL(param string, value interface{}) (url string) {
	oldValue, exists := u.values[param]

	setValue(u.values, param, value)
	url = encode(u.values, u.path)

	u.restoreValue(param, oldValue, exists)
	return
}

func (u *URLQuery) GenURLMulti2(p1 string, v1 interface{}, p2 string, v2 interface{}) (url string) {
	oldValue1, exists1 := u.values[p1]
	oldValue2, exists2 := u.values[p2]

	setValue(u.values, p1, v1)
	setValue(u.values, p2, v2)

	url = encode(u.values, u.path)

	u.restoreValue(p1, oldValue1, exists1)
	u.restoreValue(p2, oldValue2, exists2)
	return
}

func (u *URLQuery) GenURLBase(path string) string {
	//values := url.Values{"s": u.clientsString}
	return encode(u.values, path)
}

func (u *URLQuery) GenURLBaseMulti1(path string, p1 string, v1 interface{}) string {
	values := url.Values{"s": u.clientsString}
	setValue(values, p1, v1)
	return encode(values, path)
}

func (u *URLQuery) GenURLBaseMulti2(path string, p1 string, v1 interface{}, p2 string, v2 interface{}) string {
	values := url.Values{"s": u.clientsString}
	setValue(values, p1, v1)
	setValue(values, p2, v2)
	return encode(values, path)
}

func (u *URLQuery) HasCustomer(id model.CustomerID) bool {
	for _, v := range u.Customers {
		if v == id {
			return true
		}
	}
	return false
}

func (u *URLQuery) restoreValue(param string, oldValue []string, exists bool) {
	if exists {
		u.values[param] = oldValue
	} else {
		u.values.Del(param)
	}
}

func (u *URLQuery) HasSite(value database.SiteID) bool {
	for _, v := range u.Sites {
		if v == value {
			return true
		}
	}
	return false
}

func setValue(values url.Values, param string, value interface{}) {
	valueString := fmt.Sprintf("%v", value)
	if valueString == "" || valueString == "<nil>" {
		values.Del(param)
	} else {
		values.Set(param, valueString)
	}
}

func encode(values url.Values, path string) string {
	queryEncoded := values.Encode()
	if queryEncoded == "" {
		return path
	}
	return path + "?" + queryEncoded
}

func StreamPDF(file string, w http.ResponseWriter) (int, error) {
	pdfStream, err := ioutil.ReadFile(file)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	buff := bytes.NewBuffer(pdfStream)
	_, err = buff.WriteTo(w)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-type", "application/pdf")
	return http.StatusOK, nil
}

func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}
