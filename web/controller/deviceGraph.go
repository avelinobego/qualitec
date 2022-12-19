package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"celus-ti.com.br/qualitec/database/model"
	"github.com/gorilla/mux"
)

func DeviceGraph(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {

	devflag, ok := mux.Vars(r)["devflag"]
	if !ok {
		return http.StatusBadRequest, errors.New("incorrect param Devflag")
	}

	sub, err_sub := mux.Vars(r)["sub"]
	if !err_sub {
		return http.StatusBadRequest, errors.New("incorrect param sub")
	}

	device, err := model.DeviceGetByDevflag(c.DB, devflag)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	dr, err := model.DeviceViewRealTimeList(c.DB,
		model.NewViewParam(c.User.ID, c.URL.Customers),
		devflag,
		"",
		"")

	if err != nil {
		return http.StatusInternalServerError, err
	}

	//TODO Este é o struct que deverá ser criado pra represenar os dados do gráfico no formato json.
	type datasetConfig struct {
		Data        []string `json:"data"`
		Label       string   `json:"label"`
		BorderColor string   `json:"borderColor"`
		Fill        bool     `json:"fill"`
		Tension     float64  `json:"tension"`
		PointRadius int      `json:"pointRadius"`
	}
	dataset := []datasetConfig{}

	labels := []string{}

	for _, channel := range dr {

		hmodel, err := model.DeviceHistoryGetByDevflag(sub, c.DBEarth, &device.Device, channel.Channel)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		dataValues := []string{}

		for _, m := range hmodel {
			//TODO Impedir repetição
			labels = append(labels, fmt.Sprintf("%v", m.Time))
			dataValues = append(dataValues, fmt.Sprintf("%f", m.Value))
		}

		if len(dataValues) > 0 {
			dataset = append(dataset,
				datasetConfig{
					Data:        dataValues,
					Label:       channel.ChannelDescription,
					BorderColor: color(channel.GraphColor),
					Fill:        false,
					Tension:     0.1,
					PointRadius: 0,
				})
		}
	}

	type dataType struct {
		Labels   []string        `json:"labels"`
		Datasets []datasetConfig `json:"datasets"`
	}

	data := dataType{Labels: labels, Datasets: dataset}

	w.Header().Set("Content-Type", "application/json")
	valor, err := json.Marshal(data)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	w.Write(valor)

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
