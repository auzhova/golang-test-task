package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
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
	params := mux.Vars(r)
	var userId string = "1"
	if _, ok := params["user_id"]; ok {
		userId = params["user_id"]
	}

	var user User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userId)
	if err := rowUser.Scan(&user.id, &user.name); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Пользователя c id " + userId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userId, err))
		return
	}

	var balance Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userId)
	if err := rowBalance.Scan(&balance.id, &balance.user, &balance.total); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Баланс пользователя c id " + userId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userId, err))
		fmt.Sprintf(" %+v", err)
		return
	}

	var currency = map[string]float64{"RUB": 1.00, "USD": 61.25, "EUR": 61.14}

	if _, ok := params["currency"]; ok {
		balance.total = balance.total / currency[params["currency"]]
	}

	json.NewEncoder(w).Encode(balance)
	return

}

func updateBalance(w http.ResponseWriter, r *http.Request) {
	var DB = Init()
	params := mux.Vars(r)
	var userId string = "1"
	if _, ok := params["user_id"]; ok {
		userId = params["user_id"]
	}

	var user User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userId)
	if err := rowUser.Scan(&user.id, &user.name); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Пользователя c id " + userId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователя c id = %v не найден! %+v", userId, err))
		return
	}

	var balance Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userId)
	if err := rowBalance.Scan(&balance.id, &balance.user, &balance.total); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Баланс пользователя c id " + userId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userId, err))
		fmt.Sprintf(" %+v", err)
		return
	}

	var amount float64 = 0.00
	if _, ok := params["amount"]; ok {
		amount, _ = strconv.ParseFloat(params["amount"], 64)
	}

	if amount < 0 && balance.total < amount {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(fmt.Sprintf("Недостаточно средств для списания. Баланс: %f", balance.total))
		return
	}

	comment := ""

	if amount < 0 {
		comment = fmt.Sprintf("Списание суммы %f", amount)
	}

	if amount > 0 {
		comment = fmt.Sprintf("Зачисление суммы %f", amount)
	}

	balance.total = balance.total + amount
	sqlUpdateBalance := `UPDATE balances SET total = $1 WHERE id = $2;`
	rowUpdateBalance := DB.QueryRow(sqlUpdateBalance, balance.total+amount, balance.id)
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
	params := mux.Vars(r)
	var userFromId string = "1"
	if _, ok := params["user_from"]; ok {
		userFromId = params["user_from"]
	}

	var userFrom User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userFromId)
	if err := rowUser.Scan(&userFrom.id, &userFrom.name); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Пользователя c id " + userFromId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userFromId, err))
		return
	}

	var balanceFrom Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userFromId)
	if err := rowBalance.Scan(&balanceFrom.id, &balanceFrom.user, &balanceFrom.total); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Баланс пользователя c id " + userFromId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userFromId, err))
		return
	}

	var userToId string = "2"
	if _, ok := params["user_to"]; ok {
		userToId = params["user_to"]
	}

	var userTo User
	sqlUserTo := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUserTo := DB.QueryRow(sqlUserTo, userToId)
	if err := rowUserTo.Scan(&userTo.id, &userTo.name); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Пользователя c id " + userToId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userToId, err))
		return
	}

	var balanceTo Balance
	sqlBalanceTo := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalanceTo := DB.QueryRow(sqlBalanceTo, userToId)
	if err := rowBalanceTo.Scan(&balanceTo.id, &balanceTo.user, &balanceTo.total); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Баланс пользователя c id " + userToId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userToId, err))
		return
	}

	var amount float64 = 0.00
	if _, ok := params["amount"]; ok {
		amount, _ = strconv.ParseFloat(params["amount"], 64)
	}

	if balanceFrom.total < amount {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(fmt.Sprintf("Недостаточно средств для списания. Баланс: %f", balanceFrom.total))
		return
	}

	sqlUpdateBalanceFrom := `UPDATE balances SET total = $1 WHERE id = $2 RETURNING id;`
	rowUpdateBalance := DB.QueryRow(sqlUpdateBalanceFrom, balanceFrom.total-amount, balanceFrom.id)
	if err := rowUpdateBalance.Scan(&balanceFrom.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при списании с баланса %+v", err))
		return
	}

	sqlUpdateBalanceTo := `UPDATE balance SET total = $1 WHERE id = $2 RETURNING id;`
	rowUpdateBalanceTo := DB.QueryRow(sqlUpdateBalanceTo, balanceTo.total+amount, balanceTo.id)
	if err := rowUpdateBalanceTo.Scan(&balanceTo.id); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(fmt.Sprintf("Ошибка при пополнении баланса %+v", err))
		return
	}

	commentFrom := fmt.Sprintf("Перевод пользователю %s. Сумма %f", userTo.name, amount)
	commentTo := fmt.Sprintf("Перевод от пользователя %s. Сумма %f", userFrom.name, amount)

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
	params := mux.Vars(r)
	var userId string = "1"
	if _, ok := params["user_id"]; ok {
		userId = params["user_id"]
	}

	var limit string = "0"
	if _, ok := params["limit"]; ok {
		limit = params["limit"]
	}

	var ofset string = "0"
	if _, ok := params["ofset"]; ok {
		limit = params["ofset"]
	}

	var orderBy string = "ASC"
	if _, ok := params["orderBy"]; ok {
		orderBy = params["orderBy"]
	}

	var column string = "date"
	if _, ok := params["column"]; ok {
		column = params["column"]
	}

	var user User
	sqlUser := `SELECT u.id, u.name FROM users as u WHERE id=$1;`
	rowUser := DB.QueryRow(sqlUser, userId)
	if err := rowUser.Scan(&user.id, &user.name); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Пользователя c id " + userId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Пользователь c id = %v не найден! %+v", userId, err))
		return
	}

	var balance Balance
	sqlBalance := `SELECT b.id, b.user_id, b.total FROM balances as b WHERE user_id=$1;`
	rowBalance := DB.QueryRow(sqlBalance, userId)
	if err := rowBalance.Scan(&balance.id, &balance.user, &balance.total); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("Баланс пользователя c id " + userId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("Баланс пользователя c id = %v не найден! %+v", userId, err))
		return
	}

	var history History
	sqlHistory := `SELECT h.* FROM history as h WHERE balance_id=$1 ORDER BY $2 $3;`
	rowHistory := DB.QueryRow(sqlHistory, &balance.id, column, orderBy)
	if limit != "0" && ofset != "0" {
		sqlHistory = `SELECT h.* FROM history as h WHERE balance_id=$1 ORDER BY $2 $3 LIMIT $4 OFFSET $5;`
		rowHistory = DB.QueryRow(sqlHistory, &balance.id, column, orderBy, limit, ofset)
	}
	if err := rowHistory.Scan(&history.id, &history.balance); err != nil {
		w.WriteHeader(404)
		//json.NewEncoder(w).Encode("История баланса пользователя c id " + userId + " не найден!")
		json.NewEncoder(w).Encode(fmt.Sprintf("История баланса пользователя c id = %v не найдена! %+v", userId, err))
		return
	}

	json.NewEncoder(w).Encode(history)
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
	err := http.ListenAndServe(":8050", router)

	if err != nil {
		fmt.Print(err)
	}
}
