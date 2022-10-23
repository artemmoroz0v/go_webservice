package stock

import (
	t "avito/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddItem(c *gin.Context) {
	item := t.StockItem{}
	err := c.BindJSON(&item)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to parse json"})
		return
	}
	_, err = t.Db.Exec("INSERT INTO stock (product_id, product_name, product_cost, product_available) VALUES ($1, $2, $3, $4)",
		item.ProductID, item.ProductName, item.ProductCost, item.ProductAvailable)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to put item"})
		return
	}
	c.IndentedJSON(http.StatusOK, item)
}
func GetItems(c *gin.Context) {
	rows, err := t.Db.Query(("SELECT * FROM stock"))
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to get items"})
		return
	}
	defer rows.Close()
	items := []t.StockItem{}
	for rows.Next() {
		p := t.StockItem{}
		err = rows.Scan(&p.ProductID, &p.ProductName, &p.ProductCost, &p.ProductAvailable)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "problem with parsing rows"})
			return
		}
		items = append(items, p)
	}
	c.IndentedJSON(http.StatusOK, items)
}
func BuyItem(c *gin.Context) { //метод резервирования средств и метод признания выручки на счете во время покупки
	purchase := t.Purchase{}
	err := c.BindJSON(&purchase)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to parse json in BuyItem function"})
		return
	}
	//находим товар по ID и парсим его:
	item := t.StockItem{}
	request := t.Db.QueryRow("SELECT * FROM stock WHERE product_id = $1", purchase.ItemID)
	err = request.Scan(&item.ProductID, &item.ProductName, &item.ProductCost, &item.ProductAvailable)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "item with inputed id is not found"})
		return
	}
	//находим юзера по ID и парсим его:
	client := t.User{}
	request = t.Db.QueryRow("SELECT * FROM users WHERE user_id = $1", purchase.ClientID)
	err = request.Scan(&client.UserID, &client.UserName, &client.UserBalance, &client.StatusID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user with inputed id is not found"})
		return
	}
	if client.StatusID == 1 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user with 1-status can not buy anything: 1-status means that balance has been reserved"})
		return
	}
	if purchase.Price != item.ProductCost {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "please enter valid sum of wanted item"})
		return
	}
	client.StatusID = 1 //резервируем счет на момент оплаты
	if client.UserBalance >= item.ProductCost {
		if item.ProductAvailable {
			purchase.PurchaseStatus = "accepted"
			client.UserBalance -= item.ProductCost
			item.ProductAvailable = false //после покупки товар недоступен
			_, err = t.Db.Exec("UPDATE users SET user_balance = user_balance - $1 WHERE (status_id = 0) AND (user_id = $2)", item.ProductCost, client.UserID)
			if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not update user balance after purchase"})
				return
			}
			_, err = t.Db.Exec("UPDATE stock SET product_available = $1 WHERE (product_id = $2)", item.ProductAvailable, item.ProductID)
			if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not update product_available value after purchase"})
				return
			}
			client.StatusID = 0 //разрезервирование, так как все прошло успешно
			_, err = t.Db.Exec("UPDATE accounting SET spent_founds = spent_founds + $1 WHERE (accounting_id = $2)", item.ProductCost, client.UserID)
			if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not update spent_founds value after purchase"})
				return
			}
			temp := t.Accounting{}
			request = t.Db.QueryRow("SELECT * FROM accounting WHERE accounting_id = $1", client.UserID)
			_ = request.Scan(&temp.AccountingID, &temp.AccountingName, &temp.SpentFunds, &temp.Comment)
			buy_str := " Bought " + item.ProductName + " for " + strconv.Itoa(int(item.ProductCost)) + "."
			temp.Comment += buy_str
			_, err = t.Db.Exec("UPDATE accounting SET comment = $1 WHERE (accounting_id = $2)", temp.Comment, client.UserID)
			if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not update comment value after purchase"})
				return
			}
			c.IndentedJSON(http.StatusOK, gin.H{"message": "user with id " + strconv.Itoa(client.UserID) + " has bought item with id " + strconv.Itoa(item.ProductID) + " for next price: " + strconv.FormatFloat(item.ProductCost, 'f', -1, 64)})
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "product is not avaiable, your balance has been freezed"})
			purchase.PurchaseStatus = "rejected"
			_, err = t.Db.Exec("UPDATE users SET status_id = $1 WHERE (user_id = $2)", client.StatusID, client.UserID)
			if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not update user status_id to 1 because of rejected purchase"})
				return
			}
			return
		}
	} else {
		if item.ProductCost > 50000.0 {
			purchase.CanBeUnlocked = false //в случае неудачи средства могут быть разблокированы только в том случае, когда цена покупки не превышает 50 тысяч
		} else {
			purchase.CanBeUnlocked = true
		}
		purchase.PurchaseStatus = "rejected" //отказ, разрезервирования нет
		_, err = t.Db.Exec("UPDATE users SET status_id = $1 WHERE (user_id = $2)", client.StatusID, client.UserID)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not update user status_id to 1 because of rejected purchase"})
			return
		}
		_, err = t.Db.Exec("INSERT INTO failed_purchases (item_id, client_id, price, purchase_status, can_be_unlocked) VALUES ($1, $2, $3, $4, $5)", purchase.ItemID, purchase.ClientID, purchase.Price, purchase.PurchaseStatus, purchase.CanBeUnlocked)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to insert in failed_purchases"})
			return
		}
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "purchase has not been done. user's balance has been freezed!"})
	}
	c.IndentedJSON(http.StatusOK, client)
}
func UnlockBalance(c *gin.Context) {
	client := t.User{}
	err := c.BindJSON(&client)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to parse json"})
		return
	}
	rows, err := t.Db.Query(("SELECT * FROM failed_purchases"))
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to get failed_purchases"})
		return
	}
	defer rows.Close()
	purchasesArray := []t.Purchase{}
	for rows.Next() {
		temp := t.Purchase{}
		err = rows.Scan(&temp.ItemID, &temp.ClientID, &temp.Price, &temp.PurchaseStatus, &temp.CanBeUnlocked)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "problem with parsing rows in GetUsers function"})
			return
		}
		purchasesArray = append(purchasesArray, temp)
	}
	for i := 0; i < len(purchasesArray); i++ {
		if purchasesArray[i].ClientID == client.UserID {
			if client.StatusID == 0 || purchasesArray[i].PurchaseStatus == "accepted" {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user's balance is already unlocked!"})
				return
			}
			if purchasesArray[i].PurchaseStatus == "rejected" && purchasesArray[i].CanBeUnlocked {
				client.StatusID = 0
				_, err = t.Db.Exec("UPDATE users SET status_id = $1 WHERE (user_id = $2)", client.StatusID, client.UserID)
				if err != nil {
					c.IndentedJSON(http.StatusNotFound, gin.H{"message": "can not unlock user's balance"})
					return
				}
				c.IndentedJSON(http.StatusOK, gin.H{"message": "balance has been unlocked!"})
			} else {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "product cost in user's purchase was over 50 000, balance can not be unlocked"})
				return
			}
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user has not bought anything! input another id."})
			return
		}
	}
	c.IndentedJSON(http.StatusOK, client)
}

func GetAccounting(c *gin.Context) { //получить бухгалтерский отчет по юзерам
	rows, err := t.Db.Query(("SELECT * FROM accounting"))
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to get accounting results"})
		return
	}
	defer rows.Close()
	accounting := []t.Accounting{}
	for rows.Next() {
		temp := t.Accounting{}
		err = rows.Scan(&temp.AccountingID, &temp.AccountingName, &temp.SpentFunds, &temp.Comment)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "problem with parsing rows in GetAccounting function"})
			return
		}
		accounting = append(accounting, temp)
	}
	c.IndentedJSON(http.StatusOK, accounting)
}
