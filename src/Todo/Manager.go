package Todo

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

type Manager struct {
	UserNameList []string
	database     *sqlx.DB

	userCount   int64
	emptyUserId []int
	itemCount   map[int]int
	emptyItemId map[int][]int
}

func (manager *Manager) Init() {
	manager.emptyUserId = []int{}
	manager.itemCount = map[int]int{}
	manager.emptyItemId = map[int][]int{}

	db, err := sqlx.Open("mysql", "root:85864546@tcp(127.0.0.1:3306)/TODODATA")
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	manager.database = db

	err = manager.database.QueryRow("SELECT COUNT(*) FROM Users").Scan(&manager.userCount)
	if err != nil {
		log.Println("Manager.Init: Error at count users.")
	} else {
		log.Printf("Manager.Init: userCount = %v\n", manager.userCount)
	}
	manager.OutputUsers()
}

func (manager *Manager) End() {
	manager.database.Close()
}

func (manager *Manager) isUserExists(user string) bool {
	var userItems []DataBaseUserItem
	err := manager.database.Select(&userItems, "SELECT * FROM Users WHERE username = ? LIMIT 1", user)
	if err != nil {
		return false
	}
	return len(userItems) != 0
}

func (manager *Manager) isTodoItemExists(userId int, itemId int) bool {
	var count int
	err := manager.database.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ? AND id = ?", userId, itemId).Scan(&count)
	if err != nil {
		log.Printf("Manager.isTodoItemExists: Error when select from database: %v\n", err.Error())
		return false
	}
	return count != 0
}

func (manager *Manager) updateUserItemInfo(userId int) {
	_, c := manager.itemCount[userId]
	if c {
		return
	}
	var itemCount int64
	manager.database.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ?", userId).Scan(&itemCount)
	manager.itemCount[userId] = int(itemCount)
	manager.emptyItemId[userId] = make([]int, 0)
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

func (manager *Manager) AddItem(userId int, todoItem Item) (int, int) {
	var newItemId int
	var contains bool
	// Allocate item id
	emptyItemId, contains := manager.emptyItemId[userId]
	if contains || len(emptyItemId) == 0 {
		newItemId = manager.itemCount[userId]
	} else {
		newItemId = emptyItemId[0]
		manager.emptyItemId[userId] = emptyItemId[1:]
	}

	// Insert item
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

func (manager *Manager) GetItemById(userId int, itemId int) (DataBaseTodoItem, int) {
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

func (manager *Manager) GetAllItem(userId int) ([]DataBaseTodoItem, int) {
	itemList := make([]DataBaseTodoItem, 0)
	err := manager.database.Select(&itemList, "SELECT * FROM todo WHERE userid = ?", userId)
	if err != nil {
		return itemList, StatusPhraseIdError
	}
	return itemList, StatusDatabaseCommandOK
}

func (manager *Manager) UpdateItem(userId int, itemId int, values map[string]string) int {
	if !manager.isTodoItemExists(userId, itemId) {
		return StatusDatabaseSelectNotFound
	}
	command := "UPDATE todo SET "
	valueStrings := make([]string, 0)
	for key, value := range values {
		valueStrings = append(valueStrings, fmt.Sprintf("%s = %s", key, value))
	}
	log.Println(valueStrings)
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

func (manager *Manager) AddUser(user RequestLoginUserItem) int {
	var newUserId int
	if len(manager.emptyUserId) == 0 {
		newUserId = int(manager.userCount)
	} else {
		newUserId = manager.emptyUserId[0]
		manager.emptyUserId = manager.emptyUserId[1:]
	}
	_, err := manager.database.Exec("INSERT INTO Users(id, username, password) values(?, ?, ?)",
		newUserId, user.Name, toMd5(user.Password))
	if err != nil {
		log.Printf("Manager.AddUser: Error when insert user into database: %v\n", err.Error())
		return -1
	}
	manager.userCount++
	return newUserId
}

func (manager *Manager) UserLogin(user RequestLoginUserItem) (int, int) {
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

func (manager *Manager) GetUserInfo(userId int) (RequestUserInfoItem, int) {
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

	var todoCount int
	err = manager.database.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ?", userId).Scan(&todoCount)
	if err != nil {
		log.Printf("Manager.GetUserInfo: Error when select from database: %v\n", err.Error())
		return item, StatusDatabaseCommandError
	}
	item.TodoCount = todoCount

	return item, StatusDatabaseCommandOK
}

func (manager *Manager) DeleteUser(userId int) int {
	var err error
	var userItems []DataBaseUserItem
	err = manager.database.Select(&userItems, "SELECT * FROM Users WHERE id = ? LIMIT 1", userId)
	if err != nil {
		log.Println("Manager.DeleteUser: Error when select from database: ", err.Error())
		return StatusDatabaseCommandError
	}
	if len(userItems) == 0 {
		return StatusDatabaseSelectNotFound
	}

	_, err = manager.database.Exec("DELETE FROM users WHERE id = ?", userId)
	if err != nil {
		log.Println("Manager.DeleteUser: Error when delete from database: ", err.Error())
		return StatusDatabaseCommandError
	}
	_, err = manager.database.Exec("DELETE FROM todo WHERE userid = ?", userId)
	userInfo := userItems[0]
	manager.emptyUserId = append(manager.emptyUserId, userInfo.Id)
	return StatusDatabaseCommandOK
}
