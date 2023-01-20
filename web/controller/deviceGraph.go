package controller

import (
	"net/http"
)

func DeviceGraph(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {

	// c.Template.Find("graph")

	// if err := r.ParseForm(); err != nil {
	// 	return http.StatusBadRequest, errors.New("error on Parse Form")
	// }
	// f := r.Form

	// var data_device map[string]interface{}

	// if session, err := c.Store.Get(r, "session"); err == nil {
	// 	data_device = session.Values["data_device"].(map[string]interface{})
	// } else {
	// 	return http.StatusInternalServerError, err
	// }

	// deviceView := data_device["Device"].(model.DeviceView)

	// dr, err := model.DeviceViewRealTimeList(c.DB,
	// 	model.NewViewParam(c.User.ID, c.URL.Customers),
	// 	deviceView.Devflag,
	// 	"",
	// 	"")

	// if err != nil {
	// 	return http.StatusInternalServerError, err
	// }

	// type datasetConfig struct {
	// 	Data        []string
	// 	Label       string
	// 	BorderColor string
	// 	Fill        bool
	// 	Tension     float64
	// 	PointRadius int
	// }
	// dataset := []datasetConfig{}

	// labels := []string{}

	// for _, channel := range dr {

	// 	hmodel, err := model.DeviceHistoryGetByDevflag(f, c.DBEarth, &deviceView.Device, channel.Channel)
	// 	if err != nil {
	// 		return http.StatusInternalServerError, err
	// 	}

	// 	dataValues := []string{}

	// 	for _, m := range hmodel {
	// 		//TODO Impedir repetição
	// 		labels = append(labels, fmt.Sprintf("%v", m.Time))
	// 		dataValues = append(dataValues, fmt.Sprintf("%f", m.Value))
	// 	}

	// 	if len(dataValues) > 0 {
	// 		dataset = append(dataset,
	// 			datasetConfig{
	// 				Data:        dataValues,
	// 				Label:       channel.ChannelDescription,
	// 				BorderColor: color(channel.GraphColor),
	// 				Fill:        false,
	// 				Tension:     0.1,
	// 				PointRadius: 0,
	// 			})
	// 	}
	// }

	// if err := c.Template.ExecuteTemplate(w, "device-graph", data_device); err != nil {
	// 	return http.StatusInternalServerError, err
	// }

	return http.StatusOK, nil

}

var colors map[string]string = map[string]string{"red": "#ff0000", "cyan": "#00ffff", "blue": "#0000ff", "darkblue": "#00008b", "lightblue": "#add8e6", "purple": "#800080", "yellow": "#ffff00", "lime": "#00ff00", "magenta": "#ff00ff", "pink": "#ffc0cb", "white": "#ffffff", "silver": "#c0c0c0", "gray": "#808080", "black": "#000000", "orange": "#ffa500", "brown": "#a52a2a", "maroon": "#800000", "green": "#008000", "olive": "#808000", "aquamarine": "#7fffd4"}

const color_defautl = "#00ffff"

func color(name string) string {
	result := colors[name]
	if result == "" {
		result = colors[color_defautl]
	}
	return result
}
