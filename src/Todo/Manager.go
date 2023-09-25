package Todo

import (
	"TODOList/src/handler"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"strconv"
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

func (manager *Manager) isUserExists(user string) bool {
	var userItems []DataBaseUserItem
	err := manager.database.Select(&userItems, "SELECT * FROM Users WHERE username = ? LIMIT 1", user)
	if err != nil {
		return false
	}
	return len(userItems) != 0
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

func (manager *Manager) RequestRegisterUser(ctx *gin.Context) {
	var userItem RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		log.Printf("Manager.RequestRegisterUser: Error when bind json to userItem: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Error when bind json.", "token": GenerateNoUserToken()})
		return
	}
	if !isValidUsername(userItem.Name) {
		log.Printf("Manager.RequestRegisterUser: Invalid username: %v\n", userItem.Name)
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid username.", "token": GenerateNoUserToken()})
		return
	}
	if manager.isUserExists(userItem.Name) {
		log.Printf("Manager.RequestRegisterUser: User exists: %v\n", userItem.Name)
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "User exists.", "token": GenerateNoUserToken()})
		return
	}

	lastUserId := handler.GetUserIdFromToken(ctx)
	newUserId := manager.AddUser(userItem)
	if newUserId != -1 {
		ctx.JSON(http.StatusCreated,
			gin.H{"code": http.StatusCreated, "userId": newUserId, "token": GenerateUserToken(newUserId)})
		manager.updateUserItemInfo(newUserId)
		log.Printf("Manager.RequestRegisterUser: Add user successfully: %v\n", userItem.Name)
	} else {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"code": 400, "message": "Failed.", "token": GenerateUserToken(lastUserId)})
	}
}

func (manager *Manager) RequestLogin(ctx *gin.Context) {
	lastUserId := handler.GetUserIdFromToken(ctx)

	var userItem RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		log.Printf("Manager.RequestLogin: Error when bind json: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest,
			gin.H{"code": http.StatusBadRequest, "message": "Bind json error.", "token": GenerateUserToken(lastUserId)})
		return
	}
	if userItem.Name == "" {
		log.Printf("Manager.RequestLogin: User Logout: %v\n", userItem.Name)

		ctx.JSON(http.StatusOK,
			gin.H{"code": http.StatusOK, "message": "Logout successfully.", "token": GenerateNoUserToken()})
		return
	}

	userId, code := manager.UserLogin(userItem)
	if code == StatusDatabaseCommandOK {
		log.Printf("Manager.RequestLogin: User login successfully: %v\n", userItem.Name)

		ctx.JSON(http.StatusOK, gin.H{"status": "OK", "userId": code,
			"token": GenerateUserToken(userId)})
		manager.updateUserItemInfo(userId)
	} else {
		if code == StatusDatabaseSelectNotFound {
			ctx.JSON(http.StatusBadRequest,
				gin.H{"code": http.StatusBadRequest, "message": "用户名或密码错误", "token": GenerateUserToken(lastUserId)})
		} else {
			ctx.JSON(http.StatusInternalServerError,
				gin.H{"code": http.StatusInternalServerError, "message": "Error", "token": GenerateUserToken(lastUserId)})
		}
	}
}

func (manager *Manager) RequestGetCurrentUser(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "", "userinfo": gin.H{"userid": userId, "username": "", "todoCount": 0}})
		return
	}
	item, code := manager.GetUserInfo(userId)
	if code == StatusDatabaseSelectNotFound {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "User not found."})
	} else if code == StatusDatabaseCommandError {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "Internal server error."})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "", "userinfo": item})
	}
}

func (manager *Manager) RequestDeleteUser(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(http.StatusUnauthorized,
			gin.H{"code": http.StatusUnauthorized, "message": "Not login", "token": GenerateNoUserToken()})
	}
	code := manager.DeleteUser(userId)
	if code != StatusDatabaseCommandOK {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"code": http.StatusInternalServerError, "message": "Internal server error.", "token": GenerateUserToken(userId)})
		return
	}
	ctx.JSON(http.StatusOK,
		gin.H{"code": http.StatusOK, "message": "", "token": GenerateNoUserToken()})
}

func (manager *Manager) RequestAddItem(ctx *gin.Context) {
	var item RequestTodoItem
	var err error
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "Not login."})
		return
	}

	err = ctx.ShouldBindJSON(&item)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Bind json error."})
		return
	}

	itemId, code := manager.AddItem(userId, RequestToTodoItem(item))
	if code == StatusDatabaseCommandOK {
		ctx.JSON(http.StatusOK, gin.H{"status": "OK", "userId": userId, "itemId": itemId})
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
			"message": "Internal server error when insert item."})
	}
}

func (manager *Manager) RequestGetItemById(ctx *gin.Context) {
	itemId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Param id error."})
		return
	}

	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "Not login."})
		return
	}

	todoDatabaseItem, code := manager.GetItemById(userId, itemId)
	if code == StatusDatabaseCommandOK {
		requestItem := DatabaseToRequestTodoItem(todoDatabaseItem)
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "item": requestItem})
	} else if code == StatusDatabaseSelectNotFound {
		ctx.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "Item not found."})
	} else {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"code": http.StatusInternalServerError, "message": "Internal server error."})
	}
}

func (manager *Manager) RequestGetAllItem(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "Not login."})
		return
	}

	databaseItemList, code := manager.GetAllItem(userId)
	if code == StatusDatabaseCommandError {
		log.Println("Manager.RequestGetAllItem: Error when select from database.")
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"code": http.StatusInternalServerError, "message": "Error when select from database"})
		return
	}
	requestItemList := make([]RequestTodoItem, len(databaseItemList))
	for index, databaseItem := range databaseItemList {
		requestItemList[index] = DatabaseToRequestTodoItem(databaseItem)
	}
	ctx.JSON(http.StatusOK,
		gin.H{"code": http.StatusOK, "item": requestItemList})
}
