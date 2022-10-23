# Микросервис для работы с балансом пользователей.

Данный HTTP API микросервис принимает JSON и отвечает в формате JSON. Для разработки микросервиса был использован фреймворк ***gin*** и объектно-реляционная СУБД ***PostgreSQL***, а также HTTP-клиент для тестирования API ***Postman***.

### Важные примечания
1. В процессе работы был разработан не только микросервис для работы с балансом пользователей, но и целый интернет-магазин. Обдумывая реализацию обязательных методов резервирования денег, признания выручки и разблокировки баланса, а также методов разрезервирования денег в случае неудачной транзакции, получения списка транзакций для всех пользователей и для каждого по отдельности в рамках дополнительных заданий, я пришел к выводу, что было бы логично разработать некую "покупательскую" среду, то есть дать возможность нашим пользователям покупать какие-либо товары. В дальнейшем будем подробное описание каждого метода и идеи, которую я преследовал, писав тот или иной метод, чтобы получить общую картину.
2. Во время развертывания dev-среды при помощи docker и docker-compose были обнаружены неполадки: примерно в 8 из 10 случаев программа просто не могла присоединиться к базе данных, было потрачено очень много времени на то, чтобы пофиксить этот баг, но, к сожалению, не удалось: наверное, сказалось то, что пользуюсь докерскими утилитами весьма не часто. Поэтому полная и точная ее работоспособность проверена только локально. Однако изолированно система тоже может работать. Прикрепил ***docker-compose.yaml*** и ***Dockerfile*** к заданию.

Запуск: ***docker-compose up*** в командной строке, открытой в папке с заданием.

3. В проекте учитана ***каждая ошибка*** при покупках, работе с балансами и со всеми другими методами в задании. Сделано это по следующим причинам: во-первых, это очень сильно помогало отлаживать проект и устранять логические неточности; во-вторых, как было сказано, балансы пользователей - очень важные данные, которые нужно постоянно держать в тонусе. Именно поэтому на всевозможные случаи жизни в проекте предусмотрена обработка ошибок, которые также посылаются в JSON.
4. В рамках ***дополнительных заданий*** были сделаны вышеупомянутые методы разрезервирования средств пользователя и получения списка транзакций на всех пользователей и на каждого по отдельности. К сожалению, на Swagger и на написание тестов банально не хватило времени :(
5. В самом коде в файле ***main.go*** предусмотрено создание таблиц в базе данных, если они еще не созданы, однако все равно прикрепил .txt файл с SQL-запросами создания таблиц. Находится в папке ***sql_requests***.
6. Локально запустить проект можно при помощи команды ***go run main.go***.

### Описание необходимых баз данных
В рамках реализации данного проекта мне пришлось разработать ***4*** базы данных.
- база данных пользователей сервиса
- база данных товаров магазина
- база данных транзакций пользователей для бухгатерии
- база данных несостоявшихся покупок


![Screenshot](https://github.com/artemmoroz0v/go_webservice/blob/main/screenshots/Screenshot_1.png)

Разберем каждую базу данных по отдельности в деталях.
1. База данных ***пользователя*** состоит из следующих полей:
    - ***id пользователя*** - уникальный id пользователя. В базе данных это первичный ключ
    - ***имя и фамилия пользователя***
    - ***баланс пользователя***
    - ***статус пользователя*** - очень важное поле типа int. 0 означает, что средства доступны пользователю. 1 означает, что средства зарезервированы/заблокированы.

2. База данных ***товаров магазина*** состоит из следующих полей:
     - ***id товара*** - уникальный id товара. В базе данных это первичный ключ.
     - ***название товара***
     - ***стоимость товара***
     - ***доступность*** - очень важное поле типа boolean. false означает, что товар недоступен для покупки, true - доступен.

3. База данных ***транзакций пользователей** состоит из следующих полей:
     - ***id пользователя*** - первичный ключ таблицы. Внешняя ссылка на главное поле базы данных пользователя.
     - ***имя и фамилия пользователя***
     - ***количество потраченных денег*** - очень важнное поле для отчета бухгалтерии.
     - ***комментарий*** - очень важное поле, по которому видны все транзакции пользователя.

4. База данных ***несостоявшихся покупок*** cостоит из следующих полей:
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
- уменьшить средства пользователя
- перевести средства от пользователя к пользователю
- просмотреть бухгалтерский отчет как по всем сразу, так и по отдельно взятому пользователю - их историю транзакций (***доп. задание 2***)
- добавить товар для продажи
- получить список продаваемых товаров
- купить товар пользователю
- в случае блокировки баланса, разблокировать баланс пользователю (***дополнительное задание***)


### Справочник по запросам и подробные примеры работы
1. Создать нового пользователя:
    - тип запроса: ***POST**
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
