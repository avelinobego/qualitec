package controller

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"celus-ti.com.br/qualitec/database/model"
	"celus-ti.com.br/qualitec/util"
	"celus-ti.com.br/qualitec/web"
	"github.com/gorilla/mux"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

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

	c.Template.Find("device-list")

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

	err = c.Template.Execute(w, data)
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
		status, erro = graphView(dr, device, c, w, r)
	}

	return status, erro

}

func graphView(
	dr []model.DeviceViewRealTime,
	deviceView model.DeviceView,
	c *Context,
	w http.ResponseWriter,
	r *http.Request) (result int, err error) {

	result = http.StatusOK
	err = nil

	// TODO Este é o struct que deverá ser criado pra represenar os dados do gráfico no formato json.
	type datasetConfig struct {
		Values      []string `json:"data"`
		Label       string   `json:"label"`
		BorderColor string   `json:"borderColor"`
		Fill        bool     `json:"fill"`
		Tension     float64  `json:"tension"`
		PointRadius int      `json:"pointradius"`
	}

	dataset := []*datasetConfig{}

	rangeBy := strings.ToLower(r.FormValue("rangeBy"))
	dataInicial := r.FormValue("di")
	dataFinal := r.FormValue("de")

	time.Local, err = time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		result = http.StatusInternalServerError
		return
	}

	firstDay := util.FirstDate(time.Now())
	lastDay := util.LastDate(time.Now())

	if dataInicial == "" {
		dataInicial = firstDay.Format("2006-01-02")
	}

	if dataFinal == "" {
		dataFinal = lastDay.Format("2006-01-02")
	}

	if rangeBy != "" {
		dataInicial = firstDay.Format("2006-01-02")
		if rangeBy == "week" {
			lastDay = lastDay.AddDate(0, 0, 7)
		} else if rangeBy == "month" {
			lastDay = lastDay.AddDate(0, 1, 0)
		}
		dataFinal = lastDay.Format("2006-01-02")
	}

	labels := []string{}

	for _, channel := range dr {

		setup := &datasetConfig{Values: []string{},
			Label:       channel.ChannelDescription,
			BorderColor: color(channel.GraphColor),
			Fill:        false,
			Tension:     0.1,
			PointRadius: 0,
		}

		var hmodel []model.DeviceHistory
		hmodel, err = model.DeviceHistoryGetByDevflag(dataInicial, dataFinal, c.DBEarth, &deviceView.Device, channel.Channel)
		if err != nil {
			result = http.StatusInternalServerError
			return
		}

		for _, m := range hmodel {
			//TODO Impedir repetição
			evitarRepetir := make(map[string]bool)
			l := m.Time.Format("2006-01-02 15:04:05")
			if _, esta := evitarRepetir[l]; !esta {
				evitarRepetir[l] = true
				labels = append(labels, l)
			}

			evitarRepetir = make(map[string]bool)
			l = fmt.Sprintf("%f", m.Value)
			if _, esta := evitarRepetir[l]; !esta {
				evitarRepetir[l] = true
				setup.Values = append(setup.Values, l)
			}
		}

		if len(setup.Values) > 0 {
			dataset = append(dataset, setup)
		}
	}

	data_device := map[string]interface{}{
		"Labels":         labels,
		"Dataset":        dataset,
		"URL":            c.URL,
		"Device":         deviceView,
		"DeviceRealTime": dr,
		"Qdi":            dataInicial,
		"Qde":            dataFinal,
	}

	if err = c.Template.ExecuteTemplate(w, "graph", data_device); err != nil {
		result = http.StatusInternalServerError
		return
	}

	return
}

func historyView(dr []model.DeviceViewRealTime,
	device model.DeviceView,
	c *Context,
	w http.ResponseWriter,
	r *http.Request) (int, error) {

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

	c.Template.Find("pagination")

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

	err = c.Template.ExecuteTemplate(w, "history", data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
