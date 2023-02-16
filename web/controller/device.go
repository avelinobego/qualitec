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

func DeviceList(c *Context, w http.ResponseWriter, r *http.Request) (result int, err error) {

	result, err = http.StatusInternalServerError, nil

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
		var customerID model.CustomerID
		if customerID, err = model.CustomerParseID(customer_id); err == nil {
			vp.Customers = []model.CustomerID{customerID}
		} else {
			return
		}
	}

	devices, err := model.DeviceViewRealTimeList(c.DB, vp, search, like, order)
	if err != nil {
		return
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

	var customers []model.Customer
	if customers, err = model.CustomerViewListActive(c.DB, c.User.ID); err != nil {
		return
	}

	rowCount := 10
	rowInit := 1
	rowEnd := 3
	page := 1
	data := map[string]interface{}{
		"User":         c.User,
		"URL":          c.URL,
		"Customers":    customers,
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
		return
	}

	result = http.StatusOK
	err = nil

	return
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
	type dataset_struct struct {
		Label       string   `json:"label"`
		Data        []string `json:"data"`
		BorderColor string   `json:"borderColor"`
		Fill        bool     `json:"fill"`
		Tension     float64  `json:"lineTension"`
		PointRadius int      `json:"pointRadius"`
		YAxisID     string   `json:"yAxisID"`
	}

	slice_datasets := make([]*dataset_struct, 0, 1024)

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

	if rangeBy != "" {
		if rangeBy == "week" {
			firstDay = firstDay.AddDate(0, 0, -7)
		} else if rangeBy == "month" {
			firstDay = firstDay.AddDate(0, -1, 0)
		}
		dataInicial = firstDay.Format("2006-01-02")
		dataFinal = lastDay.Format("2006-01-02")
	} else {
		if dataInicial == "" || dataFinal == "" {
			dataInicial = firstDay.Format("2006-01-02")
			dataFinal = lastDay.Format("2006-01-02")
		}
	}

	labels := make([]string, 0, 1024)

	label_repetido := make(map[string]bool)

	for index, channel := range dr {

		concrete_dataset := &dataset_struct{Data: make([]string, 0, 1024),
			Label:       channel.ChannelDescription,
			BorderColor: color(channel.GraphColor),
			Fill:        false,
			Tension:     0.1,
			PointRadius: 0,
		}

		concrete_dataset.YAxisID = fmt.Sprintf("y%d", index)

		var hmodel []model.DeviceHistory
		hmodel, err = model.DeviceHistoryGetByDevflag(dataInicial,
			dataFinal,
			c.DBEarth,
			&deviceView.Device,
			channel.Channel)
		if err != nil {
			result = http.StatusInternalServerError
			return
		}

		for _, model := range hmodel {

			one_label := model.Time.Format("2006-01-02 15:04:05")
			if !label_repetido[one_label] {
				label_repetido[one_label] = true
				labels = append(labels, one_label)
			}

			one_value := fmt.Sprintf("%.4f", model.Value)
			concrete_dataset.Data = append(concrete_dataset.Data, one_value)

		}

		if len(concrete_dataset.Data) > 0 {
			slice_datasets = append(slice_datasets, concrete_dataset)
		}

	}

	data_device := map[string]interface{}{
		"Labels":         labels,
		"Dataset":        slice_datasets,
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
	r *http.Request) (result int, err error) {

	result = http.StatusInternalServerError
	err = nil

	query := r.URL.Query()

	var di time.Time
	var de time.Time
	qdi := query.Get("di")
	qde := query.Get("de")

	if qdi != "" {
		di, err = time.Parse("2006-01-02", qdi)
		if err != nil {
			return
		}
	}

	if qde != "" {
		de, err = time.Parse("2006-01-02", qde)
		if err != nil {
			return
		}
	}

	channels := make(map[string]*model.DeviceViewRealTime)
	for _, d := range dr {
		channels[d.Channel] = &d
	}

	historyCount, err := model.DeviceHistory2Count(c.DB, &device.Device, di, de, dr)
	if err != nil {
		return
	}

	linesPerPage := 30

	var page int = 1
	if qtd := query.Get("page"); qtd != "" {
		page, _ = strconv.Atoi(query.Get("page"))
	}

	if page > int(math.Ceil(float64(historyCount.Qtde)/float64(linesPerPage))) {
		page = 1
	}

	historyInit := 0
	if historyCount.Qtde > 0 {
		historyInit = linesPerPage*(page-1) + 1
	}
	historyEnd := util.Min(historyInit+linesPerPage-1, int(historyCount.Qtde))

	dh, err := model.DeviceHistory2GetByDevflag(c.DBEarth,
		&device.Device,
		dr,
		linesPerPage*(page-1),
		linesPerPage,
		di,
		de)

	if err != nil {
		return
	}

	// Pego a menor e maior data pra exibir nos campos de data
	qdi = historyCount.Qdi.Format("2006-01-02")
	qde = historyCount.Qde.Format("2006-01-02")

	c.Template.Find("pagination")

	data := map[string]interface{}{
		"Device":         device,
		"URL":            c.URL,
		"DeviceRealTime": dr,
		"DeviceHistory":  dh,
		"HistoryCount":   historyCount,
		"HistoryInit":    historyInit,
		"HistoryEnd":     historyEnd,
		"Channels":       channels,
		"Qdi":            qdi,
		"Qde":            qde,
		"Pagination": web.Pagination(c.Template, linesPerPage, historyCount.Qtde, page, func(i int) string {
			return c.URL.GenURL("page", i)
		}),
	}

	err = c.Template.ExecuteTemplate(w, "history", data)
	if err != nil {
		return
	}

	result = http.StatusOK
	err = nil

	return
}
