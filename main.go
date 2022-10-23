package main

import (
	s "avito/stock"
	t "avito/types"
	u "avito/user"
	"database/sql"
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	connection_line = "user=postgres password=postgres dbname=Avito sslmode=disable"
)

func main() {
	var err error
	t.Db, err = sql.Open("postgres", connection_line)
	if err != nil {
		log.Fatal(errors.New("error with connecting"))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(errors.New("error with closing database"))
		}
	}(t.Db)
	err = t.Db.Ping()
	if err != nil {
		log.Fatal(errors.New("error with pinging database"))
	}
	//creating databases
	_, err = t.Db.Exec("CREATE TABLE IF NOT EXISTS users (user_id INTEGER PRIMARY KEY, user_name VARCHAR (100) NOT NULL, user_balance FLOAT NOT NULL, status_id INTEGER NOT NULL)")
	if err != nil {
		log.Fatal(errors.New("error with creating database users"))
	}
	_, err = t.Db.Exec("CREATE TABLE IF NOT EXISTS stock (product_id INTEGER PRIMARY KEY, product_name VARCHAR (100) NOT NULL, product_cost FLOAT NOT NULL, product_available BOOLEAN NOT NULL)")
	if err != nil {
		log.Fatal(errors.New("error with creating database stock"))
	}
	_, err = t.Db.Exec("CREATE TABLE IF NOT EXISTS failed_purchases (item_id INTEGER PRIMARY KEY, client_id INTEGER NOT NULL, price FLOAT NOT NULL, purchase_status VARCHAR (10) NOT NULL, can_be_unlocked BOOLEAN NOT NULL, FOREIGN KEY (item_id) REFERENCES stock (product_id), FOREIGN KEY (client_id) REFERENCES users (user_id))")
	if err != nil {
		log.Fatal(errors.New("error with creating database purchases"))
	}
	_, err = t.Db.Exec("CREATE TABLE IF NOT EXISTS accounting (accounting_id INTEGER PRIMARY KEY, accounting_name VARCHAR (100) NOT NULL, spent_founds FLOAT NOT NULL, comment VARCHAR (1000) NOT NULL, FOREIGN KEY (accounting_id) REFERENCES users (user_id))")
	if err != nil {
		log.Fatal(errors.New("error with creating database accounting"))
	}
	//preparing server
	router := gin.Default()

	//users code block
	router.POST("/users/add", u.AddUser)                          //WORKS
	router.GET("/users", u.GetUsers)                              //WORKS
	router.GET("/users/:id", u.GetUserBalance)                    //WORKS
	router.PUT("/users/", u.Change)                               //WORKS
	router.PUT("/users/:fromID/:toID/:price", u.MoneyTransaction) //WORKS
	router.GET("/accounting/:id", u.GetAccountingByID)            //WORKS

	//items code block
	router.POST("/items/add", s.AddItem)         //WORKS
	router.GET("/items", s.GetItems)             //WORKS
	router.PUT("/items/buy", s.BuyItem)          //WORKS
	router.PUT("/items/unlock", s.UnlockBalance) //WORKS
	router.GET("/accounting", s.GetAccounting)   //WORKS

	//running server
	err = router.Run("localhost:8080")
	if err != nil {
		log.Fatal(errors.New("error with running router"))
	}
}
