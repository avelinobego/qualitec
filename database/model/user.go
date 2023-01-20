package model

import (
	"strconv"

	"celus-ti.com.br/qualitec/database"
	"golang.org/x/crypto/bcrypt"
)

// UserID representa o id do usuário no banco de dados
type UserID uint32

// UserParseID converte uma string contendo o id do usuário no tipo UserID
func UserParseID(id string) (UserID, error) {
	v, err := strconv.ParseUint(id, 10, 32)
	return UserID(v), err
}

// User representa um usuário no banco de dados
type User struct {
	ID        UserID
	Login     string
	Mssalkjqw string
	Email     string
	Active    bool
}

// ChangePassword altera a senha do usuário
func (u *User) ChangePassword(newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Mssalkjqw = string(hash)
	return nil
}

// IsValidPassword valida se uma determinada senha é a senha do usuário
func (u *User) IsValidPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Mssalkjqw), []byte(password))
	return err == nil
}

// Update persiste no banco de dados as modificações realizadas previamente no objeto
func (u *User) Update(db database.Exec) error {
	_, err := db.Exec("UPDATE user SET login = ?, mssalkjqw = ? WHERE id = ?", u.Login, u.Mssalkjqw, u.ID)
	return err
}

func UserExists(db database.Get, userID UserID) (bool, error) {
	count := 0
	err := db.Get(&count, "SELECT count(*) FROM user WHERE id = ?", userID)
	return count > 0, err
}

func UserGet(db database.Get, userID UserID) (user User, err error) {
	err = db.Get(&user, "SELECT id, login, mssalkjqw, email FROM user WHERE id = ?", userID)
	return
}

func UserByEmail(db database.Get, email string) (user User, err error) {
	err = db.Get(&user, "SELECT id, login, mssalkjqw, email FROM user WHERE email = ?", email)
	return
}

func UserGetAll(db database.Select) (users []User, err error) {
	err = db.Select(&users, "SELECT id, login FROM user ORDER BY login")
	return
}

// ViewParam representa os parâmetros básicos usados para filtrar
// considerando as restrições do usuário
type ViewParam struct {
	UserID    UserID
	Customers []CustomerID
}

// NewViewParam instancia um objeto ViewParam
func NewViewParam(userID UserID, customers []CustomerID) *ViewParam {
	tp := ViewParam{
		UserID:    userID,
		Customers: customers,
	}
	return &tp
}

// MakeDefWhere retorna os parâmetros com os campos "site_id" representando o ID do site
// e o campo "site_client_id" representando o ID do cliente.
func (tp *ViewParam) MakeDefWhere() (whereClauses database.Clauses, whereValues []interface{}) {
	return tp.MakeWhereWithFields("site_client_id", "site_id")
}

// MakeWhereWithFields retorna os parâmetros permitindo selecionar o nome do campo ID do
// site e o nome do campo ID do cliente. Útil para usar os filtros em outras tabelas/views
// como site_view.
func (tp *ViewParam) MakeWhereWithFields(fieldUserID, fieldSiteID string) (whereClauses database.Clauses, whereValues []interface{}) {
	whereClauses = append(whereClauses, fieldUserID+" = ?")
	whereValues = append(whereValues, tp.UserID)

	if len(tp.Customers) > 0 {
		whereClauses = append(whereClauses, fieldSiteID+" IN "+database.MakeIn(uint(len(tp.Customers))))
		s := make([]interface{}, len(tp.Customers))
		for i, v := range tp.Customers {
			s[i] = v
		}
		whereValues = append(whereValues, s...)
	}

	return

}
