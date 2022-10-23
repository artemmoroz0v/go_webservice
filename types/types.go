package types

import "database/sql"

var Db *sql.DB

type User struct {
	UserID      int     `json:"userID"`      //ник юзера
	UserName    string  `json:"userName"`    //имя-фамилия юзера
	UserBalance float64 `json:"userBalance"` // баланс юзера
	StatusID    int     `json:"statusID"`    //0 - средства доступны, 1 - средства заморожены
}

type Operation struct {
	UserID int     `json:"userID"` //id юзера
	Type   int     `json:"type"`   //тип перевода 0-перевод, 1-списание
	Value  float64 `json:"value"`  //сколько перевести
}

type Purchase struct {
	ItemID         int     `json:"itemID"`  //id товара
	ClientID       int     `json:"userID"`  //ник юзера
	Price          float64 `json:"itemSum"` //сумма товара
	PurchaseStatus string  //accepted - прошла покупка, rejected - отказано (в таком случае будет разрезервирование)
	CanBeUnlocked  bool    //может ли счет быть заблокирован
}

type StockItem struct {
	ProductID        int     `json:"productID"`        //id товара
	ProductName      string  `json:"productName"`      //название товара
	ProductCost      float64 `json:"productCost"`      //стоимость товара
	ProductAvailable bool    `json:"productAvailable"` //доступность товара
}

type Accounting struct {
	AccountingID   int     `json:"ID"`         //ник клиента
	AccountingName string  `json:"Name"`       //имя клиента
	SpentFunds     float64 `json:"spentFunds"` //потраченные средства
	Comment        string  `json:"comment"`    //комментарий о переводах средств
}
