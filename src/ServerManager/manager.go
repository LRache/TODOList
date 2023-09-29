package ServerManager

import (
	"TODOList/src/Item"
	"TODOList/src/globals"
	"TODOList/src/utils"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/wonderivan/logger"
	"gopkg.in/gomail.v2"
	"log"
	"strings"
)

type Manager struct {
	database    *sqlx.DB
	redisClient *redis.Client

	userCount int64
	itemCount map[int64]int64
}

func (manager *Manager) Init() {
	manager.itemCount = map[int64]int64{}
	var err error

	db := sqlx.MustOpen(
		globals.Configures.GetString("sql.driverName"),
		fmt.Sprintf(
			"%s:%s@tcp(%s)/%s",
			globals.Configures.GetString("sql.userName"),
			globals.Configures.GetString("sql.password"),
			globals.Configures.GetString("sql.address"),
			globals.Configures.GetString("sql.table"),
		),
	)
	if err != nil {
		logger.Alert("Manager.Init: Error when open database: ", err.Error())
		return
	}
	manager.database = db

	manager.redisClient = redis.NewClient(&redis.Options{
		Addr:     globals.Configures.GetString("redis.address"),
		Password: globals.Configures.GetString("redis.password"),
		DB:       globals.Configures.GetInt("redis.database"),
	})

	err = manager.database.QueryRow("SELECT COUNT(*) FROM Users").Scan(&manager.userCount)
	if err != nil {
		logger.Alert("Error at count users: ", err.Error())
	} else {
		logger.Trace("userCount =", manager.userCount)
	}

}

func (manager *Manager) End() {
	_ = manager.database.Close()
	_ = manager.redisClient.Close()
}

func (manager *Manager) isUserExists(mailAddr string) bool {
	var userItems []Item.DataBaseUserItem
	err := manager.database.Select(&userItems, "SELECT * FROM Users WHERE mailAddr = ? LIMIT 1", mailAddr)
	if err != nil {
		return false
	}
	return len(userItems) != 0
}

func (manager *Manager) isUserIdExists(userid int) bool {
	var userItems []Item.DataBaseUserItem
	err := manager.database.Select(&userItems, "SELECT * FROM Users WHERE id = ? LIMIT 1", userid)
	if err != nil {
		return false
	}
	return len(userItems) != 0
}

func (manager *Manager) isTodoItemExists(userId int64, itemId int64) bool {
	var count int
	err := manager.database.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ? AND id = ?", userId, itemId).Scan(&count)
	if err != nil {
		log.Printf("Manager.isTodoItemExists: Error when select from database: %v\n", err.Error())
		return false
	}
	return count != 0
}

func (manager *Manager) updateUserItemInfo(userId int64) {
	_, c := manager.itemCount[userId]
	if c {
		return
	}
	var itemCount int64
	_ = manager.database.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ?", userId).Scan(&itemCount)
	manager.itemCount[userId] = itemCount
	log.Printf("Manager.updateUserItemInfo: userid=%v itemCount=%v\n", userId, itemCount)
}

func (manager *Manager) OutputUsers() {
	var userItems []Item.DataBaseUserItem
	err := manager.database.Select(&userItems, "SELECT * FROM Users")
	if err != nil {
		logger.Error("Error at select users from database: ", err.Error())
	}
	for _, user := range userItems {
		fmt.Printf("id=%v username=%v password=%v\n", user.Id, user.Name, user.Password)
	}
}

// AddItem return new item id and result code
func (manager *Manager) AddItem(userId int64, todoItem Item.Item) (int64, int) {
	var newItemId int64
	// Allocate item id
	if manager.redisClient.LLen(fmt.Sprintf("EmptyItemId:%d", userId)).Val() == 0 {
		newItemId = manager.itemCount[userId]
	} else {
		_ = manager.redisClient.LPop(fmt.Sprintf("EmptyItemId:%d", userId)).Scan(&newItemId)
	}

	// Insert item into database
	logger.Trace("Insert item into database, userId: %v, itemId: %v", userId, newItemId)
	_, err := manager.database.Exec(
		"INSERT INTO todo(id, title, content, create_time, deadline, tag, done, userid) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		newItemId, todoItem.Title, todoItem.Content, todoItem.CreateTime, todoItem.Deadline, todoItem.Tag, todoItem.Done, userId)
	if err != nil {
		logger.Error("Error at insert item:", err.Error())
		return 0, globals.StatusDatabaseCommandError
	}
	manager.itemCount[userId]++
	return newItemId, globals.StatusDatabaseCommandOK
}

// GetItemById return item list and result code.
func (manager *Manager) GetItemById(userId int64, itemId int64) (Item.DataBaseTodoItem, int) {
	// Select item from database
	logger.Trace("(GetItemById)Select item from database, userId = %v, itemId = %v", userId, itemId)
	var todoItems []Item.DataBaseTodoItem
	err := manager.database.Select(&todoItems,
		"SELECT * FROM todo WHERE userId=? AND id=? LIMIT 1", userId, itemId)
	if err != nil {
		logger.Error("Error when select items from database: %v", err.Error())
		return Item.DataBaseTodoItem{}, globals.StatusDatabaseCommandError
	}

	if len(todoItems) == 0 {
		logger.Warn("Item not found: %v\n", itemId)
		return Item.DataBaseTodoItem{}, globals.StatusDatabaseSelectNotFound
	}
	logger.Trace("(GetItemById)Select item from database successfully, userId = %v, itemId = %v", userId, itemId)
	return todoItems[0], globals.StatusDatabaseCommandOK
}

// GetItems return item list and result code.
func (manager *Manager) GetItems(userId int64, requestItem Item.RequestGetItemsItem) ([]Item.DataBaseTodoItem, int) {
	// Generate select command.
	command := fmt.Sprintf("SELECT * FROM todo WHERE %s",
		strings.Join(append(requestItem.ToSqlSelectWhereCommandStrings(),
			fmt.Sprintf("userid = %d", userId)), " AND "))

	// Select items from database.
	logger.Trace("(GetItems)Select items from database, sqlCommand = \"%v\"", command)
	itemList := make([]Item.DataBaseTodoItem, 0)
	err := manager.database.Select(&itemList, command)
	if err != nil {
		logger.Error("(GetItems)Error when select items from database: %v", err.Error())
		return itemList, globals.StatusDatabaseCommandError
	}

	logger.Trace("(GetItems)Select items from database successfully, userId = %v, count = %v", userId, len(itemList))
	return itemList, globals.StatusDatabaseCommandOK
}

func (manager *Manager) DeleteItemById(userId int64, itemId int64) int { // Return result code.
	// Ensure item exists
	if !manager.isTodoItemExists(userId, itemId) {
		logger.Warn("(DeleteItemById)Item not exists, userId = %v, itemId = %v", userId, itemId)
		return globals.StatusDatabaseSelectNotFound
	}

	// Delete item from database
	logger.Trace("(DeleteItemById)Delete item from database: userId = %v, itemId = %v", userId, itemId)
	_, err := manager.database.Exec("DELETE FROM todo WHERE userid = ? AND id = ?", userId, itemId)
	if err != nil {
		logger.Error("(DeleteItemById)Error when delete item from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	} else {
		logger.Trace("(DeleteItemById)Delete item from database successfully: userId = %v, itemId = %v", userId, itemId)
		// Record empty item id.
		manager.redisClient.LPush(fmt.Sprintf("EmptyItemId:%d", userId), itemId)
		logger.Trace("Push empty item id to redis: userId = %v, itemId = %v", userId, itemId)
		return globals.StatusDatabaseCommandOK
	}
}

// UpdateItem return result code.
func (manager *Manager) UpdateItem(userId int64, itemId int64, values map[string]string) int {
	// Ensure item exists.
	if !manager.isTodoItemExists(userId, itemId) {
		logger.Warn("(UpdateItem)Item not exists: userId = %v, itemId = %v", userId, itemId)
		return globals.StatusDatabaseSelectNotFound
	}

	// Update item in database
	// Generate sql command
	command := "UPDATE todo SET "
	valueStrings := make([]string, 0)
	for key, value := range values {
		valueStrings = append(valueStrings, fmt.Sprintf("%s = %s", key, value))
	}
	command += strings.Join(valueStrings, ", ")
	command += fmt.Sprintf(" WHERE userid = %d AND id = %d", userId, itemId)
	logger.Trace("(UpdateItem)Update sql command: \"%s\"", command)

	_, err := manager.database.Exec(command)
	if err != nil {
		logger.Error("(UpdateItem)Error when update database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}
	logger.Trace("(UpdateItem)Update item successfully: userId = %v, itemId = %v", userId, itemId)
	return globals.StatusDatabaseCommandOK
}

// AddUser return new user id, -1 for failure.
func (manager *Manager) AddUser(user Item.RequestRegisterUserItem) int64 {
	var newUserId int64
	// Allocate new user id
	if manager.redisClient.LLen("EmptyUserId").Val() == 0 {
		newUserId = manager.userCount
	} else {
		_ = manager.redisClient.LPop("EmptyUserId").Scan(&newUserId)
	}
	logger.Trace("(AddUser)Allocate new userId = %v", newUserId)

	// Insert user into database
	logger.Trace("(AddUser)Insert new user into database: id = %v, name = %v", newUserId, user.Name)
	_, err := manager.database.Exec("INSERT INTO Users(id, username, password, todocount, mailAddr) values(?, ?, ?, 0, ?)",
		newUserId, user.Name, utils.StringToMd5(user.Password), user.MailAddr)
	if err != nil {
		logger.Error("(AddUser)Error when insert user into database: %v", err.Error())
		return -1
	}
	manager.userCount++
	logger.Trace("(AddUser)Add user successfully, userId = %v, userCount = %v", newUserId, manager.userCount)
	return newUserId
}

// UserLogin return userid and result code.
func (manager *Manager) UserLogin(user Item.RequestLoginUserItem) (int64, int) {
	// Select from database
	var userItems []Item.DataBaseUserItem
	passwordMd5 := utils.StringToMd5(user.Password)

	logger.Trace("(UserLogin)User login: mailAddr = %v, passwordMd5 = %v", user.MailAddr, passwordMd5)
	err := manager.database.Select(&userItems, "SELECT * FROM users WHERE mailAddr = ? AND password = ? LIMIT 1",
		user.MailAddr, passwordMd5)
	if err != nil {
		logger.Error("(UserLogin)Error when select user from database: %v", err.Error())
		return -1, globals.StatusDatabaseCommandError
	}

	if len(userItems) == 0 {
		logger.Warn("Manager.UserLogin: User not found: username = %v", user)
		return -1, globals.StatusDatabaseSelectNotFound
	}
	// Login successfully
	return userItems[0].Id, globals.StatusDatabaseCommandOK
}

// GetUserInfo return user info item and result code.
func (manager *Manager) GetUserInfo(userId int64) (Item.RequestUserInfoItem, int) {
	// Select user from database
	var databaseItems []Item.DataBaseUserItem
	var item Item.RequestUserInfoItem
	logger.Trace("(GetUserInfo)Select user from database: userId = %v", userId)
	err := manager.database.Select(&databaseItems, "SELECT * FROM users WHERE id = ?", userId)
	if err != nil {
		logger.Error("(GetUserInfo)Error when select from database: %v", err.Error())
		return item, globals.StatusDatabaseCommandError
	}
	if len(databaseItems) == 0 {
		logger.Warn("(GetUserInfo)User not found: userId = %v", userId)
		return item, globals.StatusDatabaseSelectNotFound
	}
	databaseItem := databaseItems[0]
	item.UserId = databaseItem.Id
	item.Name = databaseItem.Name
	item.MailAddr = databaseItem.MailAddr
	logger.Trace("(GetUserInfo)Load user item successfully: userId = %v", userId)

	// Count todoItem from database
	var todoCount int64
	logger.Trace("(GetUserInfo)Count todo from database: userId = %v", userId)
	err = manager.database.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ?", userId).Scan(&todoCount)
	if err != nil {
		logger.Error("(GetUserInfo)Error when select from database: %v", err.Error())
		return item, globals.StatusDatabaseCommandError
	}
	item.TodoCount = todoCount
	logger.Trace("(GetUserInfo)Count todo successfully: userId = %v, todoCount = %v", userId, todoCount)

	return item, globals.StatusDatabaseCommandOK
}

// DeleteUser return result code.
func (manager *Manager) DeleteUser(userId int64) int {
	var err error

	// Ensure user exists.
	var userItems []Item.DataBaseUserItem
	err = manager.database.Select(&userItems, "SELECT * FROM Users WHERE id = ? LIMIT 1", userId)
	if err != nil {
		logger.Error("(DeleteUser)Error when select from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}

	// Delete from database.
	_, err = manager.database.Exec("DELETE FROM users WHERE id = ?", userId)
	if err != nil {
		logger.Error("(DeleteUser)Error when delete user from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}
	_, err = manager.database.Exec("DELETE FROM todo WHERE userid = ?", userId)
	if err != nil {
		logger.Error("(DeleteUser)Error when delete todo items from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}

	// Record empty user id.
	manager.redisClient.LPush("EmptyUserId", userId)
	logger.Trace("(DeleteUser)Push empty userid: %v", userId)
	return globals.StatusDatabaseCommandOK
}

func (manager *Manager) SendVerifyMail(addr string) bool {
	code := utils.GenerateVerifyCode()

	m := gomail.NewMessage()
	m.SetHeader("From", globals.MailFrom)
	m.SetHeader("To", addr)
	m.SetHeader("Subject", "Verify Your Email")
	m.SetBody("text/html", fmt.Sprintf("You verify code is <br> <b>%s</b>", code))
	d := gomail.NewDialer(globals.MailServerHost, globals.MailServerPort, globals.MailSender, globals.MailPassword)
	logger.Trace("(SendVerifyMail)Send verify mail to \"%s\", verifyCode = \"%s\"", addr, code)
	err := d.DialAndSend(m)
	if err != nil {
		logger.Warn("(SendVerifyMail)Error when send mail: %v", err.Error())
		return false
	}
	manager.redisClient.HSet("MailVerifyCode", addr, code)
	logger.Trace("(SendVerifyMail)Push verify code successfully.")
	return true
}

func (manager *Manager) VerifyMail(mailAddr string, code string) (string, int) {
	cmd := manager.redisClient.HGet("MailVerifyCode", mailAddr)
	if cmd.Err() != nil {
		return "", globals.StatusNoVerifyCode
	}
	correctCode := cmd.Val()
	if correctCode == "" {
		logger.Trace("(VerifyMail)Mail not found: \"%v\"", mailAddr)
		return "", globals.StatusNoVerifyCode
	}
	if correctCode != code {
		logger.Trace("(VerifyMail)Incorrect verify code: \"%v\", correct code is \"%v\"", code, correctCode)
		return "", globals.StatusIncorrectVerifyCode
	}
	t, ok := GenerateMailVerifyCodeToken(&MailVerifyCodeClaims{MailAddr: mailAddr})
	if !ok {
		logger.Error("(VerifyMail)Error when generate token.")
		return "", globals.StatusInternalServerError
	}
	manager.redisClient.HDel("MailVerifyCode", mailAddr)
	logger.Trace("(VerifyMail)Verified mail: \"%v\".", mailAddr)
	return t, globals.StatusOK
}
