package Todo

import (
	"TODOList/src/handler"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
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
	itemId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
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
	userid := handler.GetUserIdFromToken(ctx)
	if userid == -1 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "Not login."})
		return
	}

	var requestItem RequestGetItemsItem
	err := ctx.ShouldBindQuery(&requestItem)
	if err != nil {
		log.Printf("Manager.RequestGetItems: Error when bind query: %v\n", err.Error())
		return
	}
	fmt.Println(requestItem)

	items, code := manager.GetItems(userid, requestItem)
	if code == StatusDatabaseCommandError {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"code": http.StatusInternalServerError, "message": "Internal server error."})
	} else {
		ctx.JSON(http.StatusOK,
			gin.H{"code": http.StatusOK, "message": "", "items": ListDatabaseToRequestTodoItem(items)})
	}
}

func (manager *Manager) RequestUpdateItem(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Not login.",
			})
		return
	}

	// Parse body
	var requestItem RequestUpdateTodoItem
	err := ctx.ShouldBindJSON(&requestItem)
	if err != nil {
		log.Printf("Manage.RequestUpdateItem: Error when bind json: %v\n", err.Error())
	}

	// Select items from database
	code := manager.UpdateItem(userId, requestItem.ItemId, requestItem.ToDataBaseMap())
	if code == StatusDatabaseSelectNotFound {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":    http.StatusBadRequest,
				"message": "Item not found.",
			})
	} else if code == StatusDatabaseCommandError {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    http.StatusInternalServerError,
				"message": "",
			})
	} else {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "Update successfully.",
			})
	}
}

func (manager *Manager) RequestDeleteItemById(ctx *gin.Context) {
	itemId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":    http.StatusBadRequest,
				"message": "Param error.",
			})
		return
	}

	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Not login.",
			})
		return
	}

	// Delete item from database
	code := manager.DeleteItemById(userId, itemId)
	if code == StatusDatabaseCommandOK {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "Delete item successfully.",
			})
	} else if code == StatusDatabaseSelectNotFound {
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"code":    http.StatusNotFound,
				"message": "Item not found.",
			})
	} else {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Internal server error.",
			})
	}
}

// RequestRegisterUser Return user token in json and fresh refreshToken
func (manager *Manager) RequestRegisterUser(ctx *gin.Context) {
	var userItem RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		log.Printf("Manager.RequestRegisterUser: Error when bind json to userItem: %v\n", err.Error())
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         http.StatusBadRequest,
				"message":      "Error when bind json.",
				"token":        GenerateEmptyUserToken(),
				"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}

	// Judge whether the username is valid
	if !isValidUsername(userItem.Name) {
		log.Printf("Manager.RequestRegisterUser: Invalid username: %v\n", userItem.Name)
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         400,
				"message":      "Invalid username.",
				"token":        GenerateEmptyUserToken(),
				"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}
	// Judge whether the username exists
	if manager.isUserNameExists(userItem.Name) {
		log.Printf("Manager.RequestRegisterUser: User exists: %v\n", userItem.Name)
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         400,
				"message":      "User exists.",
				"token":        GenerateEmptyUserToken(),
				"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}

	// Register
	newUserId := manager.AddUser(userItem)
	if newUserId != -1 {
		ctx.JSON(http.StatusCreated,
			gin.H{
				"code":         http.StatusCreated,
				"userId":       newUserId,
				"token":        GenerateUserToken(newUserId),
				"refreshToken": GenerateUserRefreshToken(newUserId),
			})
		manager.updateUserItemInfo(newUserId)
		log.Printf("Manager.RequestRegisterUser: Add user successfully: %v\n", userItem.Name)
	} else {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":         http.StatusInternalServerError,
				"message":      "Internal server error.",
				"token":        GetUserTokenFromContext(ctx),
				"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
			})
	}
}

// RequestLogin Return token in json and fresh refreshToken
func (manager *Manager) RequestLogin(ctx *gin.Context) {
	var userItem RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		log.Printf("Manager.RequestLogin: Error when bind json: %v\n", err.Error())
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         http.StatusBadRequest,
				"message":      "Bind json error.",
				"token":        GetUserTokenFromContext(ctx),
				"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}

	// If username is empty, logout return empty token in json.
	if userItem.Name == "" {
		log.Printf("Manager.RequestLogin: User Logout: %v\n", userItem.Name)
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":         http.StatusOK,
				"message":      "Logout successfully.",
				"token":        GenerateEmptyUserToken(),
				"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}

	// Login
	userId, code := manager.UserLogin(userItem)
	if code == StatusDatabaseCommandOK { // Login successfully
		log.Printf("Manager.RequestLogin: User login successfully: %v\n", userItem.Name)
		s := GenerateUserRefreshToken(userId)
		fmt.Println(s)
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":         http.StatusOK,
				"message":      "Login successfully.",
				"userId":       code,
				"token":        GenerateUserToken(userId),
				"refreshToken": s,
			})
		manager.updateUserItemInfo(userId)
	} else {
		if code == StatusDatabaseSelectNotFound {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{
					"code":         http.StatusUnauthorized,
					"message":      "Incorrect username or password.",
					"token":        GetUserTokenFromContext(ctx),
					"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
				})
		} else {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"code":         http.StatusInternalServerError,
					"message":      "Internal server error.",
					"token":        GetUserTokenFromContext(ctx),
					"refreshToken": GetUserRefreshTokenStringFromContext(ctx),
				})
		}
	}
}

func (manager *Manager) RequestGetCurrentUser(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)

	// User not login, return userId=-1 means no user
	if userId == -1 {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "",
				"userinfo": gin.H{
					"userid":    userId,
					"username":  "",
					"todoCount": 0,
				}})
		return
	}

	// Select user from database
	item, code := manager.GetUserInfo(userId)
	if code == StatusDatabaseSelectNotFound {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":     http.StatusBadRequest,
				"message":  "User not found.",
				"userinfo": gin.H{},
			})
	} else if code == StatusDatabaseCommandError {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":     http.StatusInternalServerError,
				"message":  "Internal server error.",
				"userinfo": gin.H{},
			})
	} else {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":     http.StatusOK,
				"message":  "",
				"userinfo": item,
			})
	}
}

func (manager *Manager) RequestDeleteUser(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Not login.",
				"token":   GenerateEmptyUserToken(),
			})
	}
	code := manager.DeleteUser(userId)
	if code != StatusDatabaseCommandOK {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Internal server error.",
				"token":   GetUserTokenFromContext(ctx),
			})
		return
	}
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"code":    http.StatusOK,
			"message": "",
			"token":   GenerateEmptyUserToken(),
		})
}

func (manager *Manager) RequestRefreshToken(ctx *gin.Context) {
	userId, b := handler.GetUserIdFromTokenIgnoreExpiration(ctx)
	// User id error
	if userId == -1 {
		if b {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"code":    http.StatusBadRequest,
					"message": "Token error.",
					"token":   GetUserTokenFromContext(ctx),
				})
		} else {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    http.StatusUnauthorized,
					"message": "Not login.",
					"token":   GetUserTokenFromContext(ctx),
				})
		}
		return
	}

	// Get refresh token id
	c := handler.ParseToken(GetUserRefreshTokenStringFromContext(ctx))
	log.Printf("Manager.RequestRefreshToken: Refresh token expires at: %v, now: %v\n",
		c.ExpiresAt, time.Now().Unix())
	// Refresh token expired
	if c.ExpiresAt < time.Now().Unix() {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Refresh token expired.",
				"token":   GetUserTokenFromContext(ctx),
			})
		return
	}
	if c.Id != userId {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid refresh token.",
				"token":   GetUserTokenFromContext(ctx),
			})
		return
	}
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"code":    http.StatusOK,
			"message": "Refresh token successfully.",
			"token":   GenerateUserToken(userId),
		})
}
