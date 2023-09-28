package ServerManager

import (
	"TODOList/src/TodoItem"
	"TODOList/src/globals"
	"TODOList/src/handler"
	"TODOList/src/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

func checkUserLogin(ctx *gin.Context) int64 {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		ctx.JSON(globals.ReturnJsonUserNotLogin.Code, globals.ReturnJsonUserNotLogin.Json)
		return -1
	}
	return userId
}

func (manager *Manager) RequestAddItem(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	var item TodoItem.RequestTodoItem
	var err error
	err = ctx.ShouldBindJSON(&item)
	if err != nil {
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
		return
	}

	itemId, code := manager.AddItem(userId, TodoItem.RequestToTodoItem(item))
	if code == globals.StatusDatabaseCommandOK {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":   http.StatusOK,
				"userId": userId,
				"itemId": itemId,
			})
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

func (manager *Manager) RequestGetItemById(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	itemId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(globals.ReturnJsonParamError.Code, globals.ReturnJsonParamError.Json)
		return
	}

	todoDatabaseItem, code := manager.GetItemById(userId, itemId)
	if code == globals.StatusDatabaseCommandOK {
		requestItem := TodoItem.DatabaseToRequestTodoItem(todoDatabaseItem)
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "item": requestItem})
	} else if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

func (manager *Manager) RequestGetItems(ctx *gin.Context) {
	userid := checkUserLogin(ctx)
	if userid == -1 {
		return
	}

	var requestItem TodoItem.RequestGetItemsItem
	err := ctx.ShouldBindQuery(&requestItem)
	if err != nil {
		log.Printf("Manager.RequestGetItems: Error when bind query: %v\n", err.Error())
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}
	fmt.Println(requestItem)

	items, code := manager.GetItems(userid, requestItem)
	if code == globals.StatusDatabaseCommandError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else {
		ctx.JSON(http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "",
				"items":   TodoItem.ListDatabaseToRequestTodoItem(items),
			})
	}
}

func (manager *Manager) RequestUpdateItem(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	// Parse body
	var requestItem TodoItem.RequestUpdateTodoItem
	err := ctx.ShouldBindJSON(&requestItem)
	if err != nil {
		log.Printf("Manage.RequestUpdateItem: Error when bind json: %v\n", err.Error())
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
	}

	// Select items from database
	code := manager.UpdateItem(userId, requestItem.ItemId, requestItem.ToDataBaseMap())
	if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else if code == globals.StatusDatabaseCommandError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	}
}

func (manager *Manager) RequestDeleteItemById(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	itemId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(globals.ReturnJsonParamError.Code, globals.ReturnJsonParamError.Json)
		return
	}

	// Delete item from database
	code := manager.DeleteItemById(userId, itemId)
	if code == globals.StatusDatabaseCommandOK {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	} else if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

// RequestRegisterUser Return user token in json and fresh refreshToken
func (manager *Manager) RequestRegisterUser(ctx *gin.Context) {
	var userItem TodoItem.RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		log.Printf("Manager.RequestRegisterUser: Error when bind json to userItem: %v\n", err.Error())
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         http.StatusBadRequest,
				"message":      "Error when bind json.",
				"token":        utils.GenerateEmptyUserToken(),
				"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}

	// Judge whether the username is valid
	if !utils.IsValidUsername(userItem.Name) {
		log.Printf("Manager.RequestRegisterUser: Invalid username: %v\n", userItem.Name)
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         http.StatusBadRequest,
				"message":      "Invalid username.",
				"token":        utils.GenerateEmptyUserToken(),
				"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}
	// Judge whether the username exists
	if manager.isUserNameExists(userItem.Name) {
		log.Printf("Manager.RequestRegisterUser: User exists: %v\n", userItem.Name)
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         http.StatusBadRequest,
				"message":      "User exists.",
				"token":        utils.GenerateEmptyUserToken(),
				"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
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
				"token":        utils.GenerateUserToken(newUserId),
				"refreshToken": utils.GenerateUserRefreshToken(newUserId),
			})
		manager.updateUserItemInfo(newUserId)
		log.Printf("Manager.RequestRegisterUser: Add user successfully: %v\n", userItem.Name)
	} else {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":         http.StatusInternalServerError,
				"message":      "Internal server error.",
				"token":        utils.GetUserTokenFromContext(ctx),
				"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
			})
	}
}

// RequestLogin Return token in json and fresh refreshToken
func (manager *Manager) RequestLogin(ctx *gin.Context) {
	var userItem TodoItem.RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		log.Printf("Manager.RequestLogin: Error when bind json: %v\n", err.Error())
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":         http.StatusBadRequest,
				"message":      "Bind json error.",
				"token":        utils.GetUserTokenFromContext(ctx),
				"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
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
				"token":        utils.GenerateEmptyUserToken(),
				"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}

	// Login
	userId, code := manager.UserLogin(userItem)
	if code == globals.StatusDatabaseCommandOK { // Login successfully
		log.Printf("Manager.RequestLogin: User login successfully: %v\n", userItem.Name)
		refreshTokenString := utils.GenerateUserRefreshToken(userId)
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":         http.StatusOK,
				"message":      "Login successfully.",
				"userId":       code,
				"token":        utils.GenerateUserToken(userId),
				"refreshToken": refreshTokenString,
			})
		manager.updateUserItemInfo(userId)
	} else {
		if code == globals.StatusDatabaseSelectNotFound {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{
					"code":         http.StatusUnauthorized,
					"message":      "Incorrect username or password.",
					"token":        utils.GetUserTokenFromContext(ctx),
					"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
				})
		} else {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"code":         http.StatusInternalServerError,
					"message":      "Internal server error.",
					"token":        utils.GetUserTokenFromContext(ctx),
					"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
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
	if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":     http.StatusBadRequest,
				"message":  "User not found.",
				"userinfo": gin.H{},
			})
	} else if code == globals.StatusDatabaseCommandError {
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
				"token":   utils.GenerateEmptyUserToken(),
			})
	}
	code := manager.DeleteUser(userId)
	if code != globals.StatusDatabaseCommandOK {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Internal server error.",
				"token":   utils.GetUserTokenFromContext(ctx),
			})
		return
	}
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"code":    http.StatusOK,
			"message": "",
			"token":   utils.GenerateEmptyUserToken(),
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
					"token":   utils.GetUserTokenFromContext(ctx),
				})
		} else {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    http.StatusUnauthorized,
					"message": "Not login.",
					"token":   utils.GetUserTokenFromContext(ctx),
				})
		}
		return
	}

	// Get refresh token id
	c := handler.ParseToken(utils.GetUserRefreshTokenStringFromContext(ctx))
	log.Printf("Manager.RequestRefreshToken: Refresh token expires at: %v, now: %v\n",
		c.ExpiresAt, time.Now().Unix())
	// Refresh token expired
	if c.ExpiresAt < time.Now().Unix() {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Refresh token expired.",
				"token":   utils.GetUserTokenFromContext(ctx),
			})
		return
	}
	if c.Id != userId {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid refresh token.",
				"token":   utils.GetUserTokenFromContext(ctx),
			})
		return
	}
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"code":    http.StatusOK,
			"message": "Refresh token successfully.",
			"token":   utils.GenerateUserToken(userId),
		})
}
