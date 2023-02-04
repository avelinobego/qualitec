package model

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"celus-ti.com.br/qualitec/database"
)

const (
	DeviceConsideredOnline = time.Hour * 6
	ModelMpm6861           = "mpm6861"
)

// DeviceID representa o id do device no banco de dados
type DeviceID uint

// DeviceParseID converte uma string contendo o id do device no tipo DeviceID
func DeviceParseID(id string) (DeviceID, error) {
	v, err := strconv.ParseUint(id, 10, 32)
	return DeviceID(v), err
}

// Device representa um device do banco de dados
type Device struct {
	ID      DeviceID
	Devflag string
	Created *time.Time
	Model   string
	Tag     string
}

// DeviceView representa a view DeviceView do banco de dados,
// que unifica os dados do device com o customer
type DeviceView struct {
	Device
	CustomerDescription  string `db:"customer_description"`
	CustomerAddress      string `db:"customer_address"`
	CustomerAddressNr    int    `db:"customer_address_nr"`
	CustomerComplement   string `db:"customer_complement"`
	CustomerNeighborhood string `db:"customer_neighborhood"`
	CustomerPhone        string `db:"customer_phone"`
	CustomerContact      string `db:"customer_contact"`
}

// DeviceGetByDevflag retorna um device pelo seu devflag
func DeviceGetByDevflag(db database.Get, devflag string) (d DeviceView, err error) {
	err = db.Get(&d,
		`SELECT
			id,
			devflag,
			model,
			IFNULL(tag, "") AS tag,
			IFNULL(customer_description, "") AS customer_description,
			IFNULL(customer_address, "") AS customer_address,
			IFNULL(customer_address_nr, 0) AS customer_address_nr,
			IFNULL(customer_complement, "") AS customer_complement,
			IFNULL(customer_neighborhood, "") AS customer_neighborhood,
			IFNULL(customer_phone, "") AS customer_phone,
			IFNULL(customer_contact, "") AS customer_contact
		FROM device_view
		WHERE devflag = ? `, devflag)
	return
}

// DeviceViewRealTime represeta a view do device no banco de dados
// que unifica e exibe de forma mais prática os dados do device
// para a tela
type DeviceViewRealTime struct {
	Device
	Customer           string
	CustomerID         int64 `db:"customer_id"`
	DecimalView        uint8 `db:"decimal_view"`
	Value              float64
	ChannelUnit        string
	ChannelDescription string
	Channel            string
	Signals            string
	Voltage            string
	GraphMin           int     `db:"graph_min"`
	GraphMax           int     `db:"graph_max"`
	GraphColor         string  `db:"graph_color"`
	ConversionFactor   float64 `db:"conversion_factor"`
	Time               *time.Time
}

// IsOnline determina se um device está online ou não com base nas configurações de hora.
// Por ora, a configuração está digitada direto no código fonte até definirmos como armazenar a configuração.
func (d *DeviceViewRealTime) IsOnline() bool {
	return d.Time.After(time.Now().UTC().Add(-DeviceConsideredOnline))
}

// DeviceViewRealTimeList retorna uma lista de devices
func DeviceViewRealTimeList(db database.Select, vp *ViewParam, devflag, like, order string) (dv []DeviceViewRealTime, err error) {
	whereClauses, whereValues := vp.MakeWhereWithFields("user_device.user_id", "device_view_realtime.customer_id")

	if like != "" {
		whereClauses = append(whereClauses, "(devflag LIKE ? OR tag LIKE ? OR customer LIKE ?)")
		searchTmp := "%" + like + "%"
		whereValues = append(whereValues, searchTmp)
		whereValues = append(whereValues, searchTmp)
		whereValues = append(whereValues, searchTmp)
	}

	if devflag != "" {
		whereClauses = append(whereClauses, "devflag = ?")
		whereValues = append(whereValues, devflag)
	}

	sqlWhere := whereClauses.MakeAndWhere(true)

	if order == "" {
		order = "customer, devflag, channel"
	}

	err = db.Select(&dv,
		`SELECT
			device_view_realtime.id,
			device_view_realtime.customer_id,
			devflag,
			model,
			IFNULL(tag, "") AS tag,
			customer,
			decimal_view,
			graph_color,
			IFNULL(graph_min, 0) AS graph_min,
			IFNULL(graph_max, 0) AS graph_max,
			IFNULL(value, 0) AS value,
			IFNULL(channel, "") AS channel,
			IFNULL(signals, "") AS signals,
			IFNULL(voltage, "") AS voltage,
			IFNULL(channel_unit, "") AS channelunit,
			IFNULL(channel_description, "") AS channeldescription,
			IFNULL(CONVERT_TZ(time, 'UTC', 'America/Sao_Paulo'), TIMESTAMP("0000-00-00")) AS time,
			conversion_factor
		FROM device_view_realtime 
		JOIN user_device ON (user_device.device_id = device_view_realtime.id)
		`+sqlWhere+` ORDER by `+order,
		whereValues...)
	return
}

// DeviceHistory representa o histórico de um device no banco dedados
// Note que ele é obtido do banco de dados do Earth1006
type DeviceHistory struct {
	Id      int64
	Channel string
	Value   float64
	Signals string
	Voltage string
	Time    *time.Time
}

const SQL_MPM6861_GRAPH string = `
SELECT
   _all.id,
   qualitec.device_channel.channel, 
   (_all.value * qualitec.device_channel.conversion_factor) AS value,
   CAST(_all.signal AS CHAR) AS signals,
   CONCAT(CAST(_all.voltage AS DECIMAL(10,2)),"v") AS voltage,
   CONVERT_TZ(_all.time, 'UTC', 'America/Sao_Paulo') AS time
FROM qualitec.device
JOIN qualitec.device_channel ON (qualitec.device_channel.device_id = qualitec.device.id)
JOIN mpm6861.chl_data_all_%[1]s _all ON (_all.channel = qualitec.device_channel.channel)
WHERE 
qualitec.device.devflag = ? AND 
qualitec.device_channel.channel = ? AND 
DATE(_all.time) BETWEEN ? AND ? 
ORDER BY _all.time 
LIMIT 100
`

const SLQ_GRAPH_EARTH = `
SELECT 
	earth.id,
	qc.channel, 
	SUBSTRING_INDEX(SUBSTRING_INDEX(earth.value, ',', FIND_IN_SET(qc.channel, earth.channel)), ',', -1) * qc.conversion_factor AS value,
	CONVERT_TZ(earth.time, 'UTC', 'America/Sao_Paulo') AS time, 
	earth.signals,
	earth.voltage
FROM qualitec.device quali
	JOIN qualitec.device_channel qc ON (qc.device_id = quali.id)
	JOIN earth1006.chl_data_prl_%[1]s earth ON earth.channel = qc.channel 
WHERE quali.devflag = ? AND qc.channel = ? 
	AND CONVERT_TZ(earth.time, 'UTC', 'America/Sao_Paulo') BETWEEN ? AND ?
	ORDER BY earth.time DESC, earth.id DESC
	LIMIT 100`

const SQL_HISTORY_MODELMPM6861 = `SELECT
		earth1006.chl_data_prl_%[1]s.id,
		qualitec.device_channel.channel,
		SUBSTRING_INDEX(SUBSTRING_INDEX(earth1006.chl_data_prl_%[1]s.value, ',', FIND_IN_SET(qualitec.device_channel.channel, earth1006.chl_data_prl_%[1]s.channel)), ',', -1) * qualitec.device_channel.conversion_factor AS value,
		CONVERT_TZ(earth1006.chl_data_prl_%[1]s.time, 'UTC', 'America/Sao_Paulo') AS time,
		earth1006.chl_data_prl_%[1]s.signals,
		earth1006.chl_data_prl_%[1]s.voltage
	FROM qualitec.device
	JOIN qualitec.device_channel ON (qualitec.device_channel.device_id = qualitec.device.id)
	JOIN earth1006.chl_data_prl_%[1]s
	WHERE qualitec.device.devflag = ? AND qualitec.device_channel.channel = ?
		AND earth1006.chl_data_prl_%[1]s.id > (SELECT CAST(MAX(earth1006.chl_data_prl_%[1]s.id) AS SIGNED) FROM earth1006.chl_data_prl_%[1]s) - 300
	ORDER BY earth1006.chl_data_prl_%[1]s.time DESC, earth1006.chl_data_prl_%[1]s.id DESC
	`

func DeviceHistoryGetByDevflag(dataInicial string, dataFinal string, db database.Select, dev *Device, channel string) (h []DeviceHistory, err error) {

	var WORK_SQL string
	var di time.Time
	var de time.Time
	di, err = time.Parse("2006-01-02", dataInicial)
	if err != nil {
		return
	}
	de, err = time.Parse("2006-01-02", dataFinal)
	if err != nil {
		return
	}

	if dev.Model == ModelMpm6861 {
		WORK_SQL = fmt.Sprintf(SQL_MPM6861_GRAPH, dev.Devflag)
	} else {
		WORK_SQL = fmt.Sprintf(SLQ_GRAPH_EARTH, dev.Devflag)
	}

	err = db.Select(&h, WORK_SQL, dev.Devflag, channel, di, de)

	return
}

/*
ALTER TABLE `device`
	ADD COLUMN `model` ENUM('6861','earth1006','mpm6861') NOT NULL DEFAULT 'mpm6861' AFTER `created`;
*/

type DeviceHistory2 struct {
	Time    *time.Time
	Channel *string
	Value   *string
	Signals *string
	Voltage *string
}

func DeviceHistory2Count(db database.Get, dev *Device, di, de *time.Time) (count int, err error) {
	if dev.Model == ModelMpm6861 {
		if di != nil && de != nil {
			sql := fmt.Sprintf(
				`SELECT count(DISTINCT time) FROM mpm6861.chl_data_all_%[1]s 
				WHERE time >= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC') AND time <= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC')
				`, dev.Devflag)
			err = db.Get(&count, sql, di, de)
		} else {
			sql := fmt.Sprintf(
				`SELECT count(DISTINCT time) FROM mpm6861.chl_data_all_%[1]s`, dev.Devflag)
			err = db.Get(&count, sql)
		}

	} else {
		if di != nil && de != nil {
			sql := fmt.Sprintf(
				`SELECT count(*) FROM earth1006.chl_data_prl_%[1]s WHERE time >= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC') AND time <= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC')`, dev.Devflag)
			err = db.Get(&count, sql, di, de)
		} else {
			sql := fmt.Sprintf(
				`SELECT count(*) FROM earth1006.chl_data_prl_%[1]s`, dev.Devflag)
			err = db.Get(&count, sql)
		}
	}
	return
}

func DeviceHistory2GetByDevflag(db database.Select,
	dev *Device,
	dr []DeviceViewRealTime,
	limitFirst,
	limitTotal int,
	di,
	de *time.Time) (dh []DeviceHistory2, err error) {

	if dev.Model == ModelMpm6861 {
		sql := bytes.NewBufferString("SELECT CONVERT_TZ(d.time, 'UTC', 'America/Sao_Paulo') AS `time`,")

		// Montagem da cláusula WHEN do SQL que faz o pivot das linhas em colunas
		var first_channel string
		last := len(dr) - 1
		values := bytes.NewBufferString("CONCAT(d.value")
		channels := bytes.NewBufferString("'")
		for index, channel := range dr {
			channels.WriteString(channel.Channel)
			if index < last {
				channels.WriteString(",")
			}
			// Pegar o primeiro canal
			if index == 0 {
				first_channel = channel.Channel
				continue
			}
			if index <= last {
				values.WriteString(",',',")
			}
			values.WriteString(fmt.Sprintf("lj%d.value", index))
		}
		values.WriteString(") as `value`, ")
		channels.WriteString("' as `channel`, ")

		sql.WriteString(channels.String())
		sql.WriteString(values.String())
		sql.WriteString("CAST(d.`signal` AS CHAR) AS `signals`, ")
		sql.WriteString("CONCAT(CAST(d.voltage AS DECIMAL(10,2)),'v') AS `voltage` ")
		sql.WriteString(fmt.Sprintf("FROM mpm6861.chl_data_all_%[1]s d ", dev.Devflag))
		for index, channel := range dr {
			// Não levo em consideração o primeiro canal
			if index == 0 {
				continue
			}
			sql.WriteString(fmt.Sprintf("LEFT JOIN  mpm6861.chl_data_all_%[1]s lj%[2]d ON lj%[2]d.channel=%[3]s AND lj%[2]d.`time` = d.`time`", dev.Devflag, index, channel.Channel))
		}

		if di != nil && de != nil {
			sql.WriteString("WHERE d.time >= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC') AND d.time <= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC') ")
			sql.WriteString("AND d.channel = '%[3]s' ")
			sql.WriteString("ORDER by d.time DESC LIMIT %[1]d, %[2]d")
			sql_final := fmt.Sprintf(sql.String(), limitFirst, limitTotal, first_channel)
			err = db.Select(&dh, sql_final, di, de)
		} else {
			sql.WriteString("WHERE d.channel = '%[3]s' ")
			sql.WriteString("ORDER by d.time DESC LIMIT %[1]d, %[2]d")
			sql_final := fmt.Sprintf(sql.String(), limitFirst, limitTotal, first_channel)
			err = db.Select(&dh, sql_final)
		}

	} else {
		if di != nil && de != nil {
			sql := fmt.Sprintf(
				`SELECT 
				channel,
				value,
				CONVERT_TZ(time, 'UTC', 'America/Sao_Paulo') AS time,
				signals,
				voltage
			FROM %[1]s.chl_data_prl_%[2]s
			WHERE time >= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC') AND time <= CONVERT_TZ(?, 'America/Sao_Paulo', 'UTC')
			ORDER by time DESC, id DESC
			LIMIT %d, %d`, dev.Model, dev.Devflag, limitFirst, limitTotal)
			err = db.Select(&dh, sql, di, de)
		} else {
			sql := fmt.Sprintf(
				`SELECT 
				channel,
				value,
				CONVERT_TZ(time, 'UTC', 'America/Sao_Paulo') AS time,
				signals,
				voltage
			FROM %[1]s.chl_data_prl_%[2]s
			ORDER by time DESC, id DESC
			LIMIT %d, %d`, dev.Model, dev.Devflag, limitFirst, limitTotal)
			err = db.Select(&dh, sql)
		}
	}

	return
}
