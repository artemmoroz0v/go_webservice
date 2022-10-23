# Микросервис для работы с балансом пользователей.

Данный HTTP API микросервис принимает JSON и отвечает в формате JSON. Для разработки микросервиса был использован микрофреймворк ***gin***, объектно-реляционная СУБД ***PostgreSQL***, а также HTTP-клиент для тестирования API ***Postman***.

### Содержание:
   - ***Важные замечания и проблемы*** : здесь я делюсь своими мыслями и проблемами 
   - ***Описание необходимых баз данных*** : здесь я объясняю, что есть что в моих базах данных
   - ***Функционал микросервиса*** : здесь я описываю функционал мироксервиса
   - ***Справочник по запросам и подробные примеры работы*** : здесь я подробно показываю работу моего проекта


### Важные примечания и прооблемы
1. В процессе работы был разработан не только микросервис для работы с балансом пользователей, но и целый интернет-магазин. Обдумывая реализацию обязательных методов резервирования денег, признания выручки и разблокировки баланса, а также методов разрезервирования денег в случае неудачной транзакции, получения списка транзакций для всех пользователей и для каждого по отдельности в рамках дополнительных заданий, я пришел к выводу, что было бы логично разработать некую "покупательскую" среду, то есть дать возможность нашим пользователям покупать какие-либо товары - выбор мой пал на покупку игровых конслей. В дальнейшем я покажу работу программы и опишу, чем я руководсвовался при реализации проекта.
2. Во время развертывания dev-среды при помощи docker и docker-compose были обнаружены неполадки: примерно в 8 из 10 случаев программа просто не могла присоединиться к базе данных, было потрачено очень много времени на то, чтобы пофиксить этот баг, но, к сожалению, не удалось: наверное, сказалось то, что пользуюсь докерскими утилитами весьма не часто. Поэтому полная и точная работоспособность веб-сервиса проверена только локально. Однако изолированно система тоже может работать. Прикрепил ***docker-compose.yaml*** и ***Dockerfile*** к заданию.

Запуск: ***docker-compose up*** в командной строке, открытой в папке с заданием.

3. В проекте учитана ***каждая ошибка*** при покупках, работе с балансами и со всеми другими методами в задании. Сделано это по следующим причинам: во-первых, это очень сильно помогало отлаживать проект и устранять логические неточности; во-вторых, как было сказано, балансы пользователей - очень важные данные, которые нужно постоянно держать в тонусе. Именно поэтому на всевозможные случаи жизни в проекте предусмотрена обработка ошибок, которые также посылаются в JSON.
4. В рамках ***дополнительных заданий*** были сделаны вышеупомянутые методы разрезервирования средств пользователя и получения списка транзакций на всех пользователей и на каждого по отдельности. К сожалению, на Swagger и на написание тестов банально не хватило времени :(
5. В самом коде в файле ***main.go*** предусмотрено создание таблиц в базе данных, если они еще не созданы, однако все равно прикрепил .txt файл с SQL-запросами создания таблиц. Находится в папке ***sql_requests***.
6. Финальные таблицы базы данных после заполнения в рамхах поясненя своего проекта в этом README-файле я прикреплю в формате .csv в папку ***final_results***.
7. Локально запустить проект можно при помощи команды ***go run main.go***. Именно в ***main.go*** идет подключение к базе данных. В случае, если нужно изменить юзера, пароль или название базы данных, можно изменить одну лишь ***const*** строку ***connection line***.

### Описание необходимых баз данных
В рамках реализации данного проекта мне пришлось разработать ***4*** базы данных.
- база данных пользователей сервиса
- база данных товаров магазина
- база данных транзакций пользователей для бухгатерии
- база данных несостоявшихся покупок

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_1.png)

Разберем каждую базу данных по отдельности в деталях.
1. База данных ***пользователя***. Думаю, тут пояснений особо не нужно. Состоит из следующих полей:
    - ***id пользователя*** - уникальный id пользователя. В базе данных это первичный ключ
    - ***имя и фамилия пользователя***
    - ***баланс пользователя***
    - ***статус пользователя*** - очень важное поле типа int. 0 означает, что средства доступны пользователю. 1 означает, что средства зарезервированы/заблокированы.

2. База данных ***товаров магазина***. Специальная база данных, чтобы пользователям было, что покупать. Состоит из следующих полей:
     - ***id товара*** - уникальный id товара. В базе данных это первичный ключ.
     - ***название товара***
     - ***стоимость товара***
     - ***доступность*** - очень важное поле типа boolean. false означает, что товар недоступен для покупки, true - доступен.

3. База данных ***транзакций пользователей***. База данных для бухгалтерии и доп.заданий. Состоит из следующих полей:
     - ***id пользователя*** - первичный ключ таблицы. Внешняя ссылка на главное поле базы данных пользователя.
     - ***имя и фамилия пользователя***
     - ***количество потраченных денег*** - очень важнное поле для отчета бухгалтерии.
     - ***комментарий*** - очень важное поле, по которому видны все транзакции пользователя.

4. База данных ***несостоявшихся покупок***. База данных для разрезервирования средств для доп.задания. Состоит из следующих полей:
     - ***id товара*** - первичный ключ таблицы. Внешняя ссылка на главное поле базы данных товаров магазина.
     - ***id пользователя*** - внешняя ссылка на главное поле базы данных пользователя.
     - ***цена товара***
     - ***статус покупки*** - отражает, совершена покупка или нет
     - ***возможность разрезервировать деньги*** очень важное поле, отражающее возможность разрезервирования денег. false в случае, если нельзя, true в случае, если можно.

### Функционал микросервиса
В нашем веб-сервисе мы можем:
- создать нового пользователя
- получить список пользователей
- получить баланс конкретного пользователя
- добавить средства пользователю
- снять средства пользователя
- перевести средства от пользователя к пользователю
- добавить товар для продажи
- получить список продаваемых товаров
- купить товар пользователю, а в случае резервирования разрезервировать ему баланс (***дополнительное задание***)
- просмотреть бухгалтерский отчет как по всем сразу, так и по отдельно взятому пользователю - их историю транзакций (***доп. задание 2***)


### Справочник по запросам и подробные примеры работы
1. Создать нового пользователя: (добавление в базу данных пользователей)
    - тип запроса: ***POST***
    - URL запроса: ***http://localhost:8080/users/add***
    - пример запроса: ***{
    "userID":0,
    "userName":"Artem Morozov",
    "userBalance":30000.0,
    "statusID":0
}***
    - ответ на запрос: ***{
    "userID":0,
    "userName":"Artem Morozov",
    "userBalance":30000.0,
    "statusID":0
}***
    - отражение в базе данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_2.png)


2. Получить список пользователей: (выборка из базы данных пользователей)
    - тип запроса: ***GET***
    - URL запроса: ***http://localhost:8080/users***
    - ответ на запрос: ***[
    {
        "userID": 0,
        "userName": "Artem Morozov",
        "userBalance": 30000,
        "statusID": 0
    },
    {
        "userID": 1,
        "userName": "Ivan Ivanov",
        "userBalance": 50000,
        "statusID": 0
    },
    {
        "userID": 2,
        "userName": "Petr Petrov",
        "userBalance": 10000,
        "statusID": 0
    },
    {
        "userID": 3,
        "userName": "Yan Yanov",
        "userBalance": 10000,
        "statusID": 0
    }
]***
    - отражение в базе данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_3.png)



3. Получить баланс пользователя. (выборка из базы данных пользователей)
    - тип запроса: ***GET***
    - URL запроса: ***http://localhost:8080/users/:id***
    - пример запроса: ***http://localhost:8080/users/0***
    - ответ на запрос: ***{
    "userID":0,
    "userName":"Artem Morozov",
    "userBalance":30000.0,
    "statusID":0
}*** 



4. Добавить средства пользователю. (изменение базы данных пользователей, добавление в учет базы данных бухгалтерии)
    - тип запроса: ***PUT***
    - URL запроса: ***http://localhost:8080/users/***
    - пример запроса: ***{
    "userID": 1,
    "type": 0,
    "value": 1000
}***
    - ответ на запрос: ***{
    "userID": 1,
    "userName": "Ivan Ivanov",
    "userBalance": 51000,
    "statusID": 0
}***
    - отражение в базе данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_4.png)


5. Снять средства пользователя (изменение базы данных пользователей, добавление в учет базы данных бухгалтерии)
    - тип запроса: ***PUT***
    - URL запроса: ***http://localhost:8080/users/***
    - пример запроса: ***{
    "userID": 1,
    "type": 1,
    "value": 5000
}***
    - ответ на запрос: ***{
    "userID": 1,
    "userName": "Ivan Ivanov",
    "userBalance": 46000,
    "statusID": 0
}***
    - отражение в базе данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_5.png)


6. Перевести средства от пользователя к пользователю (изменение базы данных пользователей, добавление в учет базы данных бухгалтерии)
    - тип запроса: ***PUT***
    - URL запроса: ***"http://localhost:8080/users/:fromID/:toID/:price"***
    - пример запроса: ***"http://localhost:8080/users/0/3/2000"***
    - ответ на запрос: ***[
    {
        "userID": 0,
        "userName": "Artem Morozov",
        "userBalance": 28000,
        "statusID": 0
    },
    {
        "userID": 3,
        "userName": "Yan Yanov",
        "userBalance": 12000,
        "statusID": 0
    }
]***
    - отражение в базе данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_6.png)

Учет пополнений, снятий и начислений выглядит в базе данных следующим образом:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_0.png)

7. Добавить товар для продажи (изменение базы данных товаров)
    - тип запроса: ***POST***
    - URL запроса: ***"http://localhost:8080/items/add"***
    - пример запроса: ***{
    "productID": 0,
    "productName": "Playstation 3",
    "productCost": 4000,
    "productAvailable": true
}***
    - ответ на запрос: ***{
    "productID": 0,
    "productName": "Playstation 3",
    "productCost": 4000,
    "productAvailable": true
}***
    - отражение в базе данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_7.png)

Вот, как выглядят эти запросы локально из кода:
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_01.png)

8. Получить список продаваемых товаров (выборка из базы данных товаров)
    - тип запроса: ***GET***
    - URL запроса: ***"http://localhost:8080/items"***
    - ответ на запрос: ***[
    {
        "productID": 0,
        "productName": "Playstation 3",
        "productCost": 4000,
        "productAvailable": true
    },
    {
        "productID": 1,
        "productName": "Playstation 4",
        "productCost": 25000,
        "productAvailable": true
    },
    {
        "productID": 2,
        "productName": "Playstation 5",
        "productCost": 65000,
        "productAvailable": true
    },
    {
        "productID": 3,
        "productName": "XBOX 360",
        "productCost": 6000,
        "productAvailable": true
    },
    {
        "productID": 4,
        "productName": "XBOX ONE",
        "productCost": 16000,
        "productAvailable": true
    },
    {
        "productID": 5,
        "productName": "XBOX SERIES S",
        "productCost": 30000,
        "productAvailable": true
    }
]***
    - отражение в базе данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_8.png)

9. Купить товар пользователю и при необходимости попытаться разрезервировать средства (изменение всех баз данных - пользователей, магазина (доступность товара меняется на false в случае покупки), транзакций (в графу соответствующего добавляются потраченные средства и комментарий, на что он их потратил), неудачных покупок (в случае отмены заказа будем видеть, можно ли разрезервировать средства или нет))

Рассмотрим два сценария: 
   - когда пользователь покупает товар успешно, 
   - когда заказ отменен, но средства можно разрезервировать
   - когда заказ отменен, но средства нельзя разрезервировать

1.  - тип запроса: ***PUT***
    - URL запроса: ***"http://localhost:8080/items/buy"***
    - пример запроса: ***{
    "itemID": 0,
    "userID": 0,
    "itemSum": 4000
}***
    - ответ на запрос: ***{
    "message": "user with id 0 has bought item with id 0 for next price: 4000"
}{
    "userID": 0,
    "userName": "Artem Morozov",
    "userBalance": 24000,
    "statusID": 0
}***
   - отражение во ***всех*** базах данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_9.png)
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_10.png)
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_11.png)
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_12.png)


2.  - тип запроса: ***PUT***
    - URL запроса: ***"http://localhost:8080/items/buy"***
    - пример запроса: ***{
    "itemID": 4,
    "userID": 2,
    "itemSum": 16000
}***
    - ответ на запрос: ***{
    "message": "purchase has not been done. user's balance has been freezed!"
}{
    "userID": 2,
    "userName": "Petr Petrov",
    "userBalance": 10000,
    "statusID": 1
}***
   - отражение во ***всех*** базах данных:

![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_13.png)
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_14.png)
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_15.png)
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_16.png)

***Самое интересное***: в базе данных ***failed_purchases*** в поле ***can_be_unlocked*** стоит значение true, так как стоимость заказа не превышала 50 тысяч. Именно такой сценарий я выбрал для реализации данного тестового задания. Если бы пользователь хотел купить заказ на сумму свыше 50 тысяч рублей, и ему бы отказало из-за недостатка средств или отсутствия товара в магазине, то разблокировать бы средства он не смог. В данном случае разрезервировать средства можно.

   - тип запроса: ***PUT***
   - URL запроса: ***"http://localhost:8080/items/unlock"***
   - пример запроса: ***{
     "userID": 2,
    "userName": "Petr Petrov",
    "userBalance": 10000,
    "statusID": 1
}***
    - ответ на запрос: ***{
    "message": "balance has been unlocked!"
}{
    "userID": 2,
    "userName": "Petr Petrov",
    "userBalance": 10000,
    "statusID": 0
}***


   - отражение в базе данных пользователя:
![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_17.png)

3. - тип запроса: ***PUT***
    - URL запроса: ***"http://localhost:8080/items/buy"***
    - пример запроса: ***{
    "itemID": 2,
    "userID": 3,
    "itemSum": 65000
}***
    - ответ на запрос: ***{
    "message": "purchase has not been done. user's balance has been freezed!"
}{
    "userID": 3,
    "userName": "Yan Yanov",
    "userBalance": 12000,
    "statusID": 1
}***

И вот тут уже при попытке разрезервировать средства вышеупомянутым запросам мы увидим следующий ответ: 

***{
    "message": "product cost in user's purchase was over 50 000, balance can not be unlocked"
}***


10. И, наконец, последнее: давайте посмотрим транзакции по отдельному пользователю и по всем пользователям.
По отдельному пользователю:
  - тип запроса: ***GET***
    - URL запроса: ***"http://localhost:8080/accounting/:id"***
    - пример запроса: ***http://localhost:8080/accounting/0***
    - ответ на запрос: ***{
    "ID": 0,
    "Name": "Artem Morozov",
    "spentFunds": 6000,
    "comment": " Transfer: 2000 to Yan Yanov. Bought Playstation 3 for 4000."
}***
    
По всем пользователям:
- тип запроса: ***GET***
    - URL запроса: ***"http://localhost:8080/accounting"***
    - пример запроса: ***http://localhost:8080/accounting***
    - ответ на запрос: ***[
    {
        "ID": 2,
        "Name": "Petr Petrov",
        "spentFunds": 0,
        "comment": ""
    },
    {
        "ID": 1,
        "Name": "Ivan Ivanov",
        "spentFunds": 0,
        "comment": " Refilled: 1000. WriteOff: 5000."
    },
    {
        "ID": 3,
        "Name": "Yan Yanov",
        "spentFunds": 0,
        "comment": " Received: 2000 from Artem Morozov."
    },
    {
        "ID": 0,
        "Name": "Artem Morozov",
        "spentFunds": 6000,
        "comment": " Transfer: 2000 to Yan Yanov. Bought Playstation 3 for 4000."
    }
]***
