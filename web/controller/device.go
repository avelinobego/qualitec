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

	/*
		dateOnline := time.Now().UTC().Add(-(model.DeviceConsideredOnline))
			// Connectivity
			connectivity := query.Get("connectivity")
			if connectivity == "online" {
				connectivity = "Online"
				whereClauses = append(whereClauses, "last_connection > ?")
				whereValues = append(whereValues, dateOnline)
			} else if connectivity == "offline" {
				connectivity = "Offline"
				whereClauses = append(whereClauses, "last_connection <= ?")
				whereValues = append(whereValues, dateOnline)
			} else {
				connectivity = ""
			}
	*/
	/*

		// status
		status := ""
		switch query.Get("status") {
		case "inuse":
			status = "Em Uso"
			whereClauses = append(whereClauses, "status = 'Em Uso'")
			break
		case "reserve":
			status = "Reserva"
			whereClauses = append(whereClauses, "status = 'Reserva'")
			break
		case "disabled":
			status = "Desativado"
			whereClauses = append(whereClauses, "status = 'Desativado'")
			break
		}

		sqlWhere := whereClauses.MakeAndWhere(true)
	*/

	// Order
	order := ""
	orderStatus := ""
	switch query.Get("order") {
	case "customer":
		order = "customer, devflag, channel"
		orderStatus = "Sorted by Customer"
		break
	case "device":
		order = "devflag, channel"
		orderStatus = "Sorted by Device"
		break
	}

	/*

		// Pagination
		rowCount := computerCount.Total
		page, err := strconv.Atoi(query.Get("page"))
		if err != nil || page > int(math.Ceil(float64(rowCount)/float64(rowsPerPage))) {
			page = 1
		}
		rowInit := 0
		if rowCount > 0 {
			rowInit = rowsPerPage*(page-1) + 1
		}
		rowEnd := util.Min(rowInit+rowsPerPage-1, rowCount)
		sqlLimit := fmt.Sprintf(" LIMIT %d, %d", rowsPerPage*(page-1), rowsPerPage)
	*/
	search := query.Get("q")

	devices, err := model.DeviceViewRealTimeList(c.DB, vp, "", search, order)
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
	/*
		computerCountSites, err := model.ComputerGetCount(c.DB, model.NewModelViewParam(c.User.ID, c.URL.Clients, c.URL.Sites))
		if err != nil {
			return http.StatusInternalServerError, err
		}
	*/

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
		//"InUseCount":    computerCount.InUse,
		//"ReserveCount":  computerCount.Reserve,
		//"DisabledCount": computerCount.Disabled,
		//"PageComputers": true,
		//"Connectivity":  connectivity,
		//"Status":        status,
		//"ComputerType":  computerTypeStr,
		"Order": orderStatus,
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

	// labels := []string{}
	// type deviceType struct {
	// 	Dataset     []string
	// 	Description string
	// 	Color       string
	// }
	// device_slice := []deviceType{}

	//ChannelDescription

	// var err error

	// for device_number, channel := range dr {

	// 	dataset := []string{}

	// 	hmodel, err := model.DeviceHistoryGetByDevflag(sub, c.DBEarth, &deviceView.Device, channel.Channel)
	// 	if err != nil {
	// 		return http.StatusInternalServerError, err
	// 	}

	// 	for _, m := range hmodel {
	// 		labels = append(labels, fmt.Sprintf("%v", m.Time))
	// 		dataset = append(dataset, fmt.Sprintf("%f", m.Value))
	// 	}

	// 	if len(dataset) > 0 {
	// 		device_slice = append(device_slice,
	// 			deviceType{
	// 				Dataset:     dataset,
	// 				Description: channel.ChannelDescription,
	// 				Color:       color(device_number),
	// 			})
	// 	}
	// }

	data := map[string]interface{}{
		// "Labels":         labels,
		// "Devices":        device_slice,
		"URL":            c.URL,
		"Device":         deviceView,
		"DeviceRealTime": dr,

		// "DeviceRealTime": dr,
		// "Graph":          graph,
		// "Data":           dataset,
	}

	// w.Header().Set("Content-Type", "application/json")
	// valor, err := json.Marshal(data)
	// if err != nil {
	// 	return http.StatusInternalServerError, err
	// }
	// w.Write(valor)

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
