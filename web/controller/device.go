package controller

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"celus-ti.com.br/qualitec/database/model"
	"celus-ti.com.br/qualitec/util"
	"celus-ti.com.br/qualitec/web"
	"github.com/gorilla/mux"
)

func DeviceList(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	rowsPerPage := 50

	query := r.URL.Query()
	vp := model.NewViewParam(c.User.ID, c.URL.Customers)

	// Order
	order := ""
	orderStatus := ""
	switch query.Get("order") {
	case "customer":
		order = "customer, devflag, channel"
		orderStatus = "Sorted by Customer"
	case "device":
		order = "devflag, channel"
		orderStatus = "Sorted by Device"
	}

	search := query.Get("q")
	like := ""

	if customer_id := query.Get("s"); customer_id != "" {
		if customerID, err := model.CustomerParseID(customer_id); err == nil {
			if result, err := model.CustomerById(c.DB, customerID); err == nil {
				like = result.Description
			} else {
				return http.StatusInternalServerError, err		
			}
		} else {
			return http.StatusInternalServerError, err
		}
	}

	devices, err := model.DeviceViewRealTimeList(c.DB, vp, search, like, order)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Contando devices online e offline
	deviceCountOnline := 0
	deviceCountOffline := 0
	devFlag := ""
	for _, d := range devices {
		if devFlag != d.Devflag {
			if d.IsOnline() {
				deviceCountOnline++
			} else {
				deviceCountOffline++
			}
			devFlag = d.Devflag
		}
	}

	rowCount := 10
	rowInit := 1
	rowEnd := 3
	page := 1
	data := map[string]interface{}{
		"User":         c.User,
		"URL":          c.URL,
		"Devices":      devices,
		"RowCount":     rowCount,
		"RowInit":      rowInit,
		"RowEnd":       rowEnd,
		"Q":            search,
		"Query":        query,
		"OnlineCount":  deviceCountOnline,
		"OfflineCount": deviceCountOffline,
		"Order":        orderStatus,
		"Pagination": web.Pagination(c.Template, rowsPerPage, rowCount, page, func(i int) string {
			return c.URL.GenURL("page", i)
		}),
	}
	err = c.Template.ExecuteTemplate(w, "device-list.html", data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// DeviceView renderiza a página de visualização do device
func DeviceView(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {

	devflag, ok := mux.Vars(r)["devflag"]
	if !ok {
		return 0, errors.New("incorrect param")
	}

	vp := model.NewViewParam(c.User.ID, c.URL.Customers)

	device, err := model.DeviceGetByDevflag(c.DB, devflag)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	dr, err := model.DeviceViewRealTimeList(c.DB, vp, devflag, "", "")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	sub := mux.Vars(r)["sub"]

	var status int = http.StatusOK
	var erro error = nil

	if sub == "history" {
		status, erro = historyView(dr, device, c, w, r)
	} else {
		status, erro = graphView(dr, device, c, w)
	}

	return status, erro

}

func graphView(
	dr []model.DeviceViewRealTime,
	deviceView model.DeviceView,
	c *Context,
	w http.ResponseWriter) (int, error) {

	data := map[string]interface{}{
		"URL":            c.URL,
		"Device":         deviceView,
		"DeviceRealTime": dr,
	}

	err := c.Template.ExecuteTemplate(w, "device-graph.html", data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func historyView(dr []model.DeviceViewRealTime, device model.DeviceView, c *Context, w http.ResponseWriter, r *http.Request) (int, error) {

	query := r.URL.Query()

	linesPerPage := 30

	var di *time.Time
	var de *time.Time
	qdi := query.Get("di")
	qde := query.Get("de")
	if qdi != "" && qde != "" {
		validation := web.NewFormValidation(r)
		di = validation.RequiredDateTimeLayout("di", 0, "02/01/2006")
		de = validation.RequiredDateTimeLayout("de", 0, "02/01/2006")
		// Como o parâmetro de só aceita data sem as horas e minutos, adiciona 23h, 59m, 59s (86339 segundos)
		// na data final (de) para que ele pegue o final do dia.
		dtmp := de.Add(time.Second * 86399)
		de = &dtmp
		if validation.Dispatch(w) {
			return http.StatusOK, nil
		}
	}

	channels := make(map[string]*model.DeviceViewRealTime)
	for _, d := range dr {
		channels[d.Channel] = &d
	}

	HistoryCount, err := model.DeviceHistory2Count(c.DB, &device.Device, di, de)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page > int(math.Ceil(float64(HistoryCount)/float64(linesPerPage))) {
		page = 1
	}

	historyInit := 0
	if HistoryCount > 0 {
		historyInit = linesPerPage*(page-1) + 1
	}
	historyEnd := util.Min(historyInit+linesPerPage-1, HistoryCount)

	dh, err := model.DeviceHistory2GetByDevflag(c.DBEarth, &device.Device, dr, linesPerPage*(page-1), linesPerPage, di, de)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	data := map[string]interface{}{
		"Device":         device,
		"URL":            c.URL,
		"DeviceRealTime": dr,
		"DeviceHistory":  dh,
		"HistoryCount":   HistoryCount,
		"HistoryInit":    historyInit,
		"HistoryEnd":     historyEnd,
		"Channels":       channels,
		"Qdi":            qdi,
		"Qde":            qde,
		"Pagination": web.Pagination(c.Template, linesPerPage, HistoryCount, page, func(i int) string {
			return c.URL.GenURL("page", i)
		}),
	}

	err = c.Template.ExecuteTemplate(w, "device-history.html", data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
