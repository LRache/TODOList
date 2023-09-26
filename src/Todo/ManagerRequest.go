package Todo

import (
	"TODOList/src/handler"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

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

func (manager *Manager) RequestGetItems(ctx *gin.Context) {
	var requestItem RequestGetItemsItem
	err := ctx.ShouldBindQuery(&requestItem)
	if err != nil {
		log.Printf("Manager.RequestGetItems: Error when bind query: %v\n", err.Error())
		return
	}
	fmt.Println(requestItem.ToSqlSelectWhereCommand())
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

func (manager *Manager) RequestUpdateItem(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(http.StatusUnauthorized,
			gin.H{"code": http.StatusUnauthorized, "message": "Not login."})
		return
	}

	var requestItem RequestUpdateTodoItem
	err := ctx.ShouldBindJSON(&requestItem)
	if err != nil {
		log.Printf("Manage.RequestUpdateItem: Error when bind json: %v\n", err.Error())
	}
	code := manager.UpdateItem(userId, requestItem.ItemId, requestItem.ToDataBaseMap())
	if code == StatusDatabaseSelectNotFound {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Item not found."})
	} else if code == StatusDatabaseCommandError {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": ""})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "Update successfully."})
	}
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
