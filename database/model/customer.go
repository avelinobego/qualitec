package model

import (
	"errors"
	"strconv"

	"celus-ti.com.br/qualitec/database"
)

// CustomerID representa o id de um cliente no banco de dados
type CustomerID uint16

// Customer representa um cliente
type Customer struct {
	ID          CustomerID
	Description string
	Active      bool
}

// CustomerParseID converte uma string contendo o id do cliente no tipo CustomerID
func CustomerParseID(id string) (CustomerID, error) {
	v, err := strconv.ParseUint(id, 10, 32)
	return CustomerID(v), err
}

// CustomerSliceStrToCustomerID convert um slice de strings contendo ids de customer
// em um slice do tipo CustomerID
func CustomerSliceStrToCustomerID(s []string, ignoreErr bool) (result []CustomerID) {
	for _, v := range s {
		converted, err := CustomerParseID(v)
		if !ignoreErr || (ignoreErr && err == nil) {
			result = append(result, converted)
		}
	}
	return
}

// CustomerViewListActive retorna uma lista view de clientes ativos ordernada pela descrição
// e filtrada pelos acessos do usuário.
func CustomerViewListActive(db database.Select, userID UserID) (s []Customer, err error) {
	err = db.Select(&s, `SELECT
		customer.id,
		customer.description
		FROM customer
		JOIN device ON (device.customer_id = customer.id)
		JOIN user_device ON (user_device.device_id = device.id)
		WHERE customer.active <> 0 AND user_device.user_id = ?
		GROUP by customer.id
		ORDER BY customer.description`, userID)
	return
}

func CustomerById(db database.Select, id CustomerID) (result *Customer, err error) {
	temp := []Customer{}
	err = db.Select(&temp, `SELECT
		customer.id,
		customer.description,
		customer.active
		FROM customer
		WHERE customer.id = ?`, id)

	if len(temp) > 1 {
		result = nil
		err = errors.New("more than one result finded")
	} else if len(temp) == 0 {
		result = nil
		err = errors.New("customer not found")
	} else {
		result = &temp[0]
	}

	return
}
