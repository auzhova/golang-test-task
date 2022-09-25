package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	id   int
	name string
}

type Balance struct {
	id    int
	user  int64
	total float64
}

type History struct {
	id      int
	balance int64
	amount  float64
	comment string
	date    time.Time
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	var DB = Init()
	params := r.URL.Query()
	var userId string = "1"
	if params.Get("user_id") != "" {
		userId = params.Get("user_id")
	}

	var user User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userId)
	if err := rowUser.Scan(&user.id, &user.name); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userId, err))
		return
	}

	var balance Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userId)
	if err := rowBalance.Scan(&balance.id, &balance.user, &balance.total); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userId, err))
		return
	}

	var currency = map[string]float64{"RUB": 1.00, "USD": 61.25, "EUR": 61.14}

	if params.Get("currency") != "" && currency[params.Get("currency")] != 0 {
		balance.total = balance.total / currency[params.Get("currency")]
	}

	response := make(map[string]string)
	response["id"] = fmt.Sprintf("%d", balance.id)
	response["user_id"] = fmt.Sprintf("%d", user.id)
	response["total"] = fmt.Sprintf("%.2f", balance.total)

	json.NewEncoder(w).Encode(response)
	return

}

func updateBalance(w http.ResponseWriter, r *http.Request) {
	var DB = Init()
	var params map[string]interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(body, &params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userId string = "1"
	if _, ok := params["user_id"]; ok {
		userId = fmt.Sprintf("%v", params["user_id"])
	}

	var user User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userId)
	if err := rowUser.Scan(&user.id, &user.name); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователя c id = %v не найден! %+v", userId, err))
		return
	}

	var balance Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userId)
	if err := rowBalance.Scan(&balance.id, &balance.user, &balance.total); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userId, err))
		fmt.Sprintf(" %+v", err)
		return
	}

	var amount float64 = 0.00
	if _, ok := params["amount"]; ok {
		amount, _ = strconv.ParseFloat(fmt.Sprintf("%v", params["amount"]), 64)
	}

	if amount < 0 && balance.total < amount {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(fmt.Sprintf("Недостаточно средств для списания. Баланс: %.2f", balance.total))
		return
	}

	comment := ""

	if amount < 0 {
		comment = fmt.Sprintf("Списание суммы %.2f", amount)
	}

	if amount > 0 {
		comment = fmt.Sprintf("Зачисление суммы %.2f", amount)
	}

	balance.total = balance.total + amount
	sqlUpdateBalance := `UPDATE balances SET total = $1 WHERE id = $2 RETURNING id;`
	rowUpdateBalance := DB.QueryRow(sqlUpdateBalance, balance.total, balance.id)
	if err := rowUpdateBalance.Scan(&balance.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при обновлении баланса %+v", err))
		return
	}

	var history History
	createdAt := time.Now()
	sqlHistory := `INSERT INTO history(balance_id, amount, comment, date) VALUES ($1, $2, $3, $4) RETURNING id;`
	rowHistory := DB.QueryRow(sqlHistory, balance.id, amount, comment, createdAt)
	if err := rowHistory.Scan(&history.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при сохранении действия в историю %+v", err))
		return
	}

	json.NewEncoder(w).Encode("Баланс успешно обновлен")
	return
}

func transferBalance(w http.ResponseWriter, r *http.Request) {
	var DB = Init()
	var params map[string]interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(body, &params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userFromId string = "1"
	if _, ok := params["user_from"]; ok {
		userFromId = fmt.Sprintf("%v", params["user_from"])
	}

	var userFrom User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userFromId)
	if err := rowUser.Scan(&userFrom.id, &userFrom.name); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userFromId, err))
		return
	}

	var balanceFrom Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userFromId)
	if err := rowBalance.Scan(&balanceFrom.id, &balanceFrom.user, &balanceFrom.total); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userFromId, err))
		return
	}

	var userToId string = "2"
	if _, ok := params["user_to"]; ok {
		userToId = fmt.Sprintf("%v", params["user_to"])
	}

	var userTo User
	sqlUserTo := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUserTo := DB.QueryRow(sqlUserTo, userToId)
	if err := rowUserTo.Scan(&userTo.id, &userTo.name); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userToId, err))
		return
	}

	var balanceTo Balance
	sqlBalanceTo := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalanceTo := DB.QueryRow(sqlBalanceTo, userToId)
	if err := rowBalanceTo.Scan(&balanceTo.id, &balanceTo.user, &balanceTo.total); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userToId, err))
		return
	}

	var amount float64 = 0.00
	if _, ok := params["amount"]; ok {
		amount, _ = strconv.ParseFloat(fmt.Sprintf("%v", params["amount"]), 64)
	}

	if balanceFrom.total < amount {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(fmt.Sprintf("Недостаточно средств для списания. Баланс: %.2f", balanceFrom.total))
		return
	}

	sqlUpdateBalanceFrom := `UPDATE balances SET total = $1 WHERE id = $2 RETURNING id;`
	rowUpdateBalance := DB.QueryRow(sqlUpdateBalanceFrom, balanceFrom.total-amount, balanceFrom.id)
	if err := rowUpdateBalance.Scan(&balanceFrom.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при списании с баланса %+v", err))
		return
	}

	sqlUpdateBalanceTo := `UPDATE balances SET total = $1 WHERE id = $2 RETURNING id;`
	rowUpdateBalanceTo := DB.QueryRow(sqlUpdateBalanceTo, balanceTo.total+amount, balanceTo.id)
	if err := rowUpdateBalanceTo.Scan(&balanceTo.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при пополнении баланса %+v", err))
		return
	}

	commentFrom := fmt.Sprintf("Перевод пользователю %s. Сумма %.2f", userTo.name, amount)
	commentTo := fmt.Sprintf("Перевод от пользователя %s. Сумма %.2f", userFrom.name, amount)

	var historyFrom History
	createdAt := time.Now()
	sqlHistoryFrom := `INSERT INTO history(balance_id, amount, comment, date) VALUES ($1, $2, $3, $4) RETURNING id;`
	rowHistoryFrom := DB.QueryRow(sqlHistoryFrom, balanceFrom.id, amount, commentFrom, createdAt)
	if err := rowHistoryFrom.Scan(&historyFrom.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при сохранении действия в историю %+v", err))
		return
	}

	var historyTo History
	sqlHistoryTo := `INSERT INTO history(balance_id, amount, comment, date) VALUES ($1, $2, $3, $4) RETURNING id;`
	rowHistoryTo := DB.QueryRow(sqlHistoryTo, balanceTo.id, amount, commentTo, createdAt)
	if err := rowHistoryTo.Scan(&historyTo.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при сохранении действия в историю %+v", err))
		return
	}

	json.NewEncoder(w).Encode("Перевод успешно выполнен")
	return

}

func historyBalance(w http.ResponseWriter, r *http.Request) {
	var DB = Init()
	params := r.URL.Query()
	var userId string = "1"
	if params.Get("user_id") != "" {
		userId = params.Get("user_id")
	}

	var limit string = ""
	if params.Get("limit") != "" {
		limit = params.Get("limit")
	}

	var offset string = ""
	if params.Get("offset") != "" {
		offset = params.Get("offset")
	}

	var orderBy string = "ASC"
	if params.Get("order_by") != "" {
		orderBy = params.Get("order_by")
	}

	var column string = "date"
	if params.Get("column") != "" {
		column = params.Get("column")
	}

	var user User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userId)
	if err := rowUser.Scan(&user.id, &user.name); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userId, err))
		return
	}

	var balance Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userId)
	if err := rowBalance.Scan(&balance.id, &balance.user, &balance.total); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userId, err))
		return
	}

	sqlHistory := `SELECT h.* FROM history as h WHERE balance_id=$1 ORDER BY '$2' $3;`
	rows, err := DB.Query(sqlHistory, &balance.id, column, orderBy)
	if limit != "" && offset != "" {
		sqlHistory = `SELECT h.* FROM history as h WHERE balance_id=$1 ORDER BY '$2' $3 LIMIT $4 OFFSET $5;`
		//rows, err := DB.Query(sqlHistory, &balance.id, column, orderBy, limit, offset)
	}

	if err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(fmt.Sprintf("История баланса пользователя c id = %v не найдена! %+v", userId, err))
		return
	}

	type Response struct {
		id         string
		balance_id string
		amount     string
		comment    string
		date       string
	}

	responses := make([]*Response, 0)
	for rows.Next() {
		var history History
		response := new(Response)
		rows.Scan(&history.id, &history.balance, &history.amount, &history.comment, &history.date)
		response.id = fmt.Sprintf("%d", history.id)
		response.balance_id = fmt.Sprintf("%d", balance.id)
		response.amount = fmt.Sprintf("%.2f", history.amount)
		response.comment = history.comment
		response.date = history.date.String()
		responses = append(responses, response)
	}
	log.Println(responses)
	json.NewEncoder(w).Encode(responses)
	return
}

func main() {
	router := mux.NewRouter()

	// GET Получение баланса текущего пользователя, с учетом курса ?currency=USD
	router.HandleFunc("/balance", getBalance).Methods("GET")
	// PUT Обновление баланса (зачисление/списание). user_id, amount
	router.HandleFunc("/balance", updateBalance).Methods("PATCH")
	// POST Перевод средств между пользователями user_from, user_to, amount
	router.HandleFunc("/balance/transfer", transferBalance).Methods("POST")
	// GET user_id, column, orderBy, limit, offset
	router.HandleFunc("/balance/history", historyBalance).Methods("GET")

	//Запустите приложение, посетите localhost:8090
	err := http.ListenAndServe(":8090", router)

	if err != nil {
		fmt.Print(err)
	}
}
