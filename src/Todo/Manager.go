package Todo

import (
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

type Manager struct {
	UserNameList []string
	database     *sqlx.DB
	redisClient  *redis.Client

	userCount int64
	itemCount map[int64]int64
}

func (manager *Manager) Init() {
	manager.itemCount = map[int64]int64{}

	db, err := sqlx.Open("mysql", "root:85864546@tcp(127.0.0.1:3306)/TODODATA")
	if err != nil {
		log.Panicf("Manager.Init: Error when open database: %v", err.Error())
		return
	}
	manager.database = db

	manager.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	err = manager.database.QueryRow("SELECT COUNT(*) FROM Users").Scan(&manager.userCount)
	if err != nil {
		log.Printf("Manager.Init: Error at count users.")
	} else {
		log.Printf("Manager.Init: userCount=%v", manager.userCount)
	}

}

func (manager *Manager) End() {
	_ = manager.database.Close()
}

func (manager *Manager) isUserNameExists(user string) bool {
	var userItems []DataBaseUserItem
	err := manager.database.Select(&userItems, "SELECT * FROM Users WHERE username = ? LIMIT 1", user)
	if err != nil {
		return false
	}
	return len(userItems) != 0
}

func (manager *Manager) isUserIdExists(userid int) bool {
	var userItems []DataBaseUserItem
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
	var userItems []DataBaseUserItem
	err := manager.database.Select(&userItems, "SELECT * FROM Users")
	if err != nil {
		fmt.Println("Error at select users from database: ", err.Error())
	}
	for _, user := range userItems {
		fmt.Printf("id=%v username=%v password=%v\n", user.Id, user.Name, user.Password)
	}
}

func (manager *Manager) AddItem(userId int64, todoItem Item) (int64, int) {
	var newItemId int64
	// Allocate item id
	if manager.redisClient.LLen(fmt.Sprintf("EmptyItemId:%d", userId)).Val() == 0 {
		newItemId = manager.itemCount[userId]
	} else {
		_ = manager.redisClient.LPop(fmt.Sprintf("EmptyItemId:%d", userId)).Scan(&newItemId)
	}

	// Insert item into database
	_, err := manager.database.Exec(
		"INSERT INTO todo(id, title, content, create_time, deadline, tag, done, userid) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		newItemId, todoItem.Title, todoItem.Content, todoItem.CreateTime, todoItem.Deadline, todoItem.Tag, todoItem.Done, userId)
	if err != nil {
		log.Printf("Manager.AddItem: Error at insert item: %v", err.Error())
		return 0, StatusDatabaseCommandError
	}
	manager.itemCount[userId]++
	return newItemId, StatusDatabaseCommandOK
}

func (manager *Manager) GetItemById(userId int64, itemId int64) (DataBaseTodoItem, int) {
	var todoItems []DataBaseTodoItem
	err := manager.database.Select(&todoItems,
		"SELECT * FROM todo WHERE userId=? AND id=? LIMIT 1", userId, itemId)
	if err != nil {
		log.Printf("Manager.GetItemById: Error when select items from database: %v\n", err.Error())
		return DataBaseTodoItem{}, StatusDatabaseCommandError
	}
	if len(todoItems) == 0 {
		log.Printf("Manager.GetItemById: Item not found: %v\n", itemId)
		return DataBaseTodoItem{}, StatusDatabaseSelectNotFound
	}
	return todoItems[0], StatusDatabaseCommandOK
}

func (manager *Manager) GetItems(userId int64, requestItem RequestGetItemsItem) ([]DataBaseTodoItem, int) {
	itemList := make([]DataBaseTodoItem, 0)
	command := fmt.Sprintf("SELECT * FROM todo WHERE %s",
		strings.Join(append(requestItem.ToSqlSelectWhereCommandStrings(),
			fmt.Sprintf("userid = %d", userId)), " AND "))
	fmt.Println(command)
	err := manager.database.Select(&itemList, command)
	if err != nil {
		return itemList, StatusDatabaseCommandError
	}
	return itemList, StatusDatabaseCommandOK
}

func (manager *Manager) DeleteItemById(userId int64, itemId int64) int {
	// Ensure item exists
	if !manager.isTodoItemExists(userId, itemId) {
		return StatusDatabaseSelectNotFound
	}

	// Delete item from database
	_, err := manager.database.Exec("DELETE FROM todo WHERE userid = ? AND id = ?", userId, itemId)
	if err != nil {
		return StatusDatabaseCommandError
	} else {
		// Record empty item id.
		manager.redisClient.LPush(fmt.Sprintf("EmptyItemId:%d", userId), itemId)
		return StatusDatabaseCommandOK
	}
}

func (manager *Manager) UpdateItem(userId int64, itemId int64, values map[string]string) int {
	// Ensure item exists.
	if !manager.isTodoItemExists(userId, itemId) {
		return StatusDatabaseSelectNotFound
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
	log.Printf("Manager.UpdateItem: Sql command: \n%s\n", command)

	_, err := manager.database.Exec(command)
	if err != nil {
		log.Printf("Manage.UpdateItem: Error when update database: %v\n", err.Error())
		return StatusDatabaseCommandError
	}
	return StatusDatabaseCommandOK
}

func (manager *Manager) AddUser(user RequestLoginUserItem) int64 {
	var newUserId int64
	if manager.redisClient.LLen("EmptyUserId").Val() == 0 {
		newUserId = manager.userCount
	} else {
		_ = manager.redisClient.LPop("EmptyUserId").Scan(&newUserId)
	}
	_, err := manager.database.Exec("INSERT INTO Users(id, username, password, todocount) values(?, ?, ?, 0)",
		newUserId, user.Name, toMd5(user.Password))
	if err != nil {
		log.Printf("Manager.AddUser: Error when insert user into database: %v\n", err.Error())
		return -1
	}
	manager.userCount++
	return newUserId
}

func (manager *Manager) UserLogin(user RequestLoginUserItem) (int64, int) {
	var userItems []DataBaseUserItem
	passwordMd5 := toMd5(user.Password)
	err := manager.database.Select(&userItems, "SELECT * FROM users WHERE username = ? AND password = ? LIMIT 1",
		user.Name, passwordMd5)
	if err != nil {
		log.Printf("Manager.UserLogin: Error when select user from database: %v\n", err.Error())
		return -1, StatusDatabaseCommandError
	}
	if len(userItems) == 0 {
		log.Printf("Manager.UserLogin: User not found: %v\n", user)
		return -1, StatusDatabaseSelectNotFound
	}
	return userItems[0].Id, StatusDatabaseCommandOK
}

func (manager *Manager) GetUserInfo(userId int64) (RequestUserInfoItem, int) {
	var databaseItems []DataBaseUserItem
	var item RequestUserInfoItem

	err := manager.database.Select(&databaseItems, "SELECT * FROM users WHERE id = ?", userId)
	if err != nil {
		log.Printf("Manager.GetUserInfo: Error when select from database: %v\n", err.Error())
		return item, StatusDatabaseCommandError
	}
	if len(databaseItems) == 0 {
		return item, StatusDatabaseSelectNotFound
	}
	databaseItem := databaseItems[0]
	item.UserId = databaseItem.Id
	item.Name = databaseItem.Name

	var todoCount int64
	err = manager.database.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ?", userId).Scan(&todoCount)
	if err != nil {
		log.Printf("Manager.GetUserInfo: Error when select from database: %v\n", err.Error())
		return item, StatusDatabaseCommandError
	}
	item.TodoCount = todoCount

	return item, StatusDatabaseCommandOK
}

func (manager *Manager) DeleteUser(userId int64) int {
	var err error

	// Ensure user exists.
	var userItems []DataBaseUserItem
	err = manager.database.Select(&userItems, "SELECT * FROM Users WHERE id = ? LIMIT 1", userId)
	if err != nil {
		log.Printf("Manager.DeleteUser: Error when select from database: %v", err.Error())
		return StatusDatabaseCommandError
	}

	// Delete from database.
	_, err = manager.database.Exec("DELETE FROM users WHERE id = ?", userId)
	if err != nil {
		log.Printf("Manager.DeleteUser: Error when delete from database: %v", err.Error())
		return StatusDatabaseCommandError
	}
	_, err = manager.database.Exec("DELETE FROM todo WHERE userid = ?", userId)
	if err != nil {
		return StatusDatabaseCommandError
	}

	// Record empty user id.
	userInfo := userItems[0]
	manager.redisClient.LPush("EmptyUserId", userInfo.Id)
	return StatusDatabaseCommandOK
}
