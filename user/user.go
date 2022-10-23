package user

import (
	t "avito/types"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddUser(c *gin.Context) {
	client := t.User{}
	err := c.BindJSON(&client)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to parse json"})
		return
	}
	if client.StatusID == 1 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "inputed user can not have 1 status id!"})
		return
	}
	//добавляем юзера в базу данных users
	_, err = t.Db.Exec("INSERT INTO users (user_id, user_name, user_balance, status_id) VALUES ($1, $2, $3, $4)", client.UserID, client.UserName, client.UserBalance, client.StatusID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to put user"})
		return
	}
	//добавляем юзера в бухгалтерский учет
	_, err = t.Db.Exec("INSERT INTO accounting (accounting_id, accounting_name, spent_founds, comment) VALUES ($1, $2, $3, $4)", client.UserID, client.UserName, 0.0, "")
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to add user in accounting database in AddUser function"})
		return
	}
	c.IndentedJSON(http.StatusOK, client)
}
func GetUsers(c *gin.Context) {
	rows, err := t.Db.Query(("SELECT * FROM users"))
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to get users"})
		return
	}
	defer rows.Close()
	users := []t.User{}
	for rows.Next() {
		temp := t.User{}
		err = rows.Scan(&temp.UserID, &temp.UserName, &temp.UserBalance, &temp.StatusID)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "problem with parsing rows in GetUsers function"})
			return
		}
		users = append(users, temp)
	}
	c.IndentedJSON(http.StatusOK, users)
}
func GetUserBalance(c *gin.Context) {
	id := c.Param("id")
	client := t.User{}
	request := t.Db.QueryRow("SELECT * FROM users WHERE user_id = $1", id)
	err := request.Scan(&client.UserID, &client.UserName, &client.UserBalance, &client.StatusID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user with inputed id is not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, client)
}
func RefillBalance(client t.User, id int, value float64) error { //зачисление средств
	if client.StatusID == 1 {
		return errors.New("can not refill balance")
	}
	//пополнение средств
	_, err := t.Db.Exec("UPDATE users SET user_balance = user_balance + $1 WHERE (status_id = 0) AND (user_id = $2)",
		value, id)
	if err != nil {
		return errors.New("can not refill balance")
	}
	temp := t.Accounting{}
	request := t.Db.QueryRow("SELECT * FROM accounting WHERE accounting_id = $1", id)
	_ = request.Scan(&temp.AccountingID, &temp.AccountingName, &temp.SpentFunds, &temp.Comment)
	refill_str := " Refilled: " + strconv.Itoa(int(value)) + "."
	temp.Comment += refill_str
	_, err = t.Db.Exec("UPDATE accounting SET comment = $1 WHERE (accounting_id = $2)", temp.Comment, id)
	if err != nil {
		return errors.New("problem with updating comment in accounting database")
	}
	return nil
}
func WriteOffBalance(client t.User, id int, value float64) error { //списание средств
	if client.StatusID == 1 {
		return errors.New("can not write off balance (status_id = 1")
	}
	if client.UserBalance < value {
		return errors.New("unable to write off balance (client balance is less than the inputed value)")
	}
	_, err := t.Db.Exec("UPDATE users SET user_balance = user_balance - $1 WHERE (status_id = 0) AND (user_id = $2)", value, id)
	if err != nil {
		return errors.New("can not writeoff balance")
	}
	temp := t.Accounting{}
	request := t.Db.QueryRow("SELECT * FROM accounting WHERE accounting_id = $1", id)
	_ = request.Scan(&temp.AccountingID, &temp.AccountingName, &temp.SpentFunds, &temp.Comment)
	writeoff_str := " WriteOff: " + strconv.Itoa(int(value)) + "."
	temp.Comment += writeoff_str
	_, err = t.Db.Exec("UPDATE accounting SET comment = $1 WHERE (accounting_id = $2)", temp.Comment, id)
	if err != nil {
		return errors.New("problem with updating comment in accounting database")
	}
	return nil
}
func Change(c *gin.Context) { //функция операция
	client := t.User{}
	op := t.Operation{}
	err := c.BindJSON(&op)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to parse operation"})
		return
	}
	request := t.Db.QueryRow("SELECT * FROM users WHERE user_id = $1", op.UserID)
	_ = request.Scan(&client.UserID, &client.UserName, &client.UserBalance, &client.StatusID)
	if client.StatusID == 1 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "client's status is 1 - you can not refill or write off money"})
		return
	}
	switch op.Type {
	case 0:
		err = RefillBalance(client, op.UserID, op.Value)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		}
		request = t.Db.QueryRow("SELECT * FROM users WHERE user_id = $1", op.UserID)
		_ = request.Scan(&client.UserID, &client.UserName, &client.UserBalance, &client.StatusID)
		c.IndentedJSON(http.StatusOK, client)
	case 1:
		err = WriteOffBalance(client, op.UserID, op.Value)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user's balance is less than the inputed data, balance can not be negative"})
		} else {
			request := t.Db.QueryRow("SELECT * FROM users WHERE user_id = $1", op.UserID)
			_ = request.Scan(&client.UserID, &client.UserName, &client.UserBalance, &client.StatusID)
			c.IndentedJSON(http.StatusOK, client)
		}
	default:
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "operation type is only 0 or 1"})
	}
}
func MoneyTransaction(c *gin.Context) {
	fromID := c.Param("fromID")
	toID := c.Param("toID")
	price_str := c.Param("price")
	price_float64, err := strconv.ParseFloat(price_str, 64)
	if err != nil {
		log.Fatal(errors.New("can not convert string to float64"))
	}
	clientFrom := t.User{}
	clientTo := t.User{}
	request := t.Db.QueryRow("SELECT * FROM users WHERE user_id = $1", fromID)
	err = request.Scan(&clientFrom.UserID, &clientFrom.UserName, &clientFrom.UserBalance, &clientFrom.StatusID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user that wanted to transfer money is not found"})
		return
	}
	request = t.Db.QueryRow("SELECT * FROM users WHERE user_id = $1", toID)
	err = request.Scan(&clientTo.UserID, &clientTo.UserName, &clientTo.UserBalance, &clientTo.StatusID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user to whom is needed to transfer money is not found"})
		return
	}
	if clientFrom.UserBalance < price_float64 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "not enough money for transfer"})
		return
	} else if clientFrom.StatusID == 1 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not transfer money, sender's balance had been freezed"})
		return
	} else if clientTo.StatusID == 1 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not transfer money, receiver's balance had been freezed"})
		return
	}
	_, err = t.Db.Exec("UPDATE users SET user_balance = user_balance - $1 WHERE (status_id = 0) AND (user_id = $2)", price_float64, clientFrom.UserID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "something wrong with database decrement exec has been happened"})
		return
	}
	clientFrom.UserBalance -= price_float64
	_, err = t.Db.Exec("UPDATE accounting SET spent_founds = spent_founds + $1 WHERE (accounting_id = $2)", price_float64, clientFrom.UserID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "something wrong with adding in accounting database in MoneyTransaction"})
		return
	}
	tempfrom := t.Accounting{}
	request = t.Db.QueryRow("SELECT * FROM accounting WHERE accounting_id = $1", fromID)
	_ = request.Scan(&tempfrom.AccountingID, &tempfrom.AccountingName, &tempfrom.SpentFunds, &tempfrom.Comment)
	transfer_string := " Transfer: " + price_str + " to " + clientTo.UserName + "."
	tempfrom.Comment += transfer_string
	_, err = t.Db.Exec("UPDATE accounting SET comment = $1 WHERE (accounting_id = $2)", tempfrom.Comment, clientFrom.UserID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "something wrong with adding comment in accounting database in MoneyTransaction"})
		return
	}
	tempto := t.Accounting{}
	request = t.Db.QueryRow("SELECT * FROM accounting WHERE accounting_id = $1", toID)
	_ = request.Scan(&tempto.AccountingID, &tempto.AccountingName, &tempto.SpentFunds, &tempto.Comment)
	receiver_string := " Received: " + price_str + " from " + clientFrom.UserName + "."
	tempto.Comment += receiver_string
	_, err = t.Db.Exec("UPDATE accounting SET comment = $1 WHERE (accounting_id = $2)", tempto.Comment, clientTo.UserID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "something wrong with adding comment in accounting database in MoneyTransaction"})
		return
	}
	_, err = t.Db.Exec("UPDATE users SET user_balance = user_balance + $1 WHERE (status_id = 0) AND (user_id = $2)", price_float64, clientTo.UserID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "something wrong with database increment exec has been happened"})
		return
	}
	clientTo.UserBalance += price_float64
	jsonUsersArray := []t.User{clientFrom, clientTo}
	c.IndentedJSON(http.StatusOK, jsonUsersArray)
}

func GetAccountingByID(c *gin.Context) {
	ID := c.Param("id")
	client := t.Accounting{}
	request := t.Db.QueryRow("SELECT * FROM accounting WHERE accounting_id = $1", ID)
	err := request.Scan(&client.AccountingID, &client.AccountingName, &client.SpentFunds, &client.Comment)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not get accounting by id!"})
		return
	}
	c.IndentedJSON(http.StatusOK, client)
}
