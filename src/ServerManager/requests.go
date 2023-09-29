package ServerManager

import (
	"TODOList/src/Item"
	"TODOList/src/globals"
	"TODOList/src/handler"
	"TODOList/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// checkUserLogin return -1 if user not login, and set context.
func checkUserLogin(ctx *gin.Context) int64 {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		logger.Warn("User not login.")
		ctx.JSON(globals.ReturnJsonUserNotLogin.Code, globals.ReturnJsonUserNotLogin.Json)
		return -1
	}
	return userId
}

// RequestAddItem send new item id.
func (manager *Manager) RequestAddItem(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	var item Item.RequestTodoItem
	var err error
	err = ctx.ShouldBindJSON(&item)
	if err != nil {
		logger.Warn("(RequestAddItem)Bind body json error: %v", err.Error())
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
		return
	}

	itemId, code := manager.AddItem(userId, Item.RequestToTodoItem(item))
	if code == globals.StatusDatabaseCommandOK {
		ctx.JSON(
			http.StatusCreated,
			gin.H{
				"code":    http.StatusCreated,
				"message": "",
				"userId":  userId,
				"itemId":  itemId,
			})
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

// RequestGetItemById send a item using RequestTodoItem type.
func (manager *Manager) RequestGetItemById(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	itemId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("(RequestGetItemById)Error when parse param: %v", err.Error())
		ctx.JSON(globals.ReturnJsonParamError.Code, globals.ReturnJsonParamError.Json)
		return
	}

	todoDatabaseItem, code := manager.GetItemById(userId, itemId)
	if code == globals.StatusDatabaseCommandOK {
		requestItem := Item.DatabaseToRequestTodoItem(todoDatabaseItem)
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "item": requestItem})
	} else if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

// RequestGetItems send item list using RequestTodoItem type.
func (manager *Manager) RequestGetItems(ctx *gin.Context) {
	userid := checkUserLogin(ctx)
	if userid == -1 {
		return
	}

	var requestItem Item.RequestGetItemsItem
	err := ctx.ShouldBindQuery(&requestItem)
	if err != nil {
		logger.Warn("(RequestGetItems)Error when bind query: %v", err.Error())
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}

	items, code := manager.GetItems(userid, requestItem)
	if code == globals.StatusDatabaseCommandError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else {
		ctx.JSON(http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "",
				"items":   Item.ListDatabaseToRequestTodoItem(items),
			})
	}
}

// RequestUpdateItem send code and message.
func (manager *Manager) RequestUpdateItem(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	// Parse body
	var requestItem Item.RequestUpdateTodoItem
	err := ctx.ShouldBindJSON(&requestItem)
	if err != nil {
		logger.Warn("(RequestUpdateItem)Error when bind body json: %v", err.Error())
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
		logger.Warn("(RequestGetItemById)Error when parse param: %v", err.Error())
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

// RequestRegisterUser send user token in json and fresh refreshToken
func (manager *Manager) RequestRegisterUser(ctx *gin.Context) {
	var userItem Item.RequestRegisterUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		logger.Warn("(RequestRegisterUser)Error when bind body json to userItem: %v", err.Error())
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

	// Judge whether the username exists
	if manager.isUserExists(userItem.MailAddr) {
		logger.Info("(RequestRegisterUser)User exists: %v", userItem.Name)
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

	if !strings.HasSuffix(userItem.MailAddr, "@todouser") {
		mailInToken, ok := GetMailAddrFromToken(userItem.MailToken)
		if !ok {
			logger.Warn("(RequestRegisterUser)Get mail from token failed.")
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"code":         http.StatusBadRequest,
					"message":      "Parse token failed.",
					"token":        utils.GenerateEmptyUserToken(),
					"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
				})
			return
		} else if mailInToken != userItem.MailAddr {
			logger.Trace("(RequestRegisterUser)Unmatched token, mailInToken = %v but got %v",
				mailInToken, userItem.MailAddr)
			ctx.JSON(
				http.StatusNotAcceptable,
				gin.H{
					"code":         http.StatusNotAcceptable,
					"message":      "Unmatched mail.",
					"token":        utils.GenerateEmptyUserToken(),
					"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
				})
			return
		}
	}

	// Judge whether the username is valid
	if !utils.IsValidUsername(userItem.Name) {
		logger.Info("(RequestRegisterUser)Invalid username: %v", userItem.Name)
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

// RequestLogin send token in json and fresh refresh token.
func (manager *Manager) RequestLogin(ctx *gin.Context) {
	var userItem Item.RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		logger.Warn("Manager.RequestLogin: Error when bind json: %v", err.Error())
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

	// If userAddr is empty, logout return empty token in json.
	if userItem.MailAddr == "" {
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
		logger.Trace("Manager.RequestLogin: User login successfully: %v", userItem.MailAddr)
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
		logger.Error("(RequestGetCurrentUser)User not login.")
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "",
				"userinfo": gin.H{
					"userid":    userId,
					"username":  "",
					"todoCount": 0,
					"mailAddr":  "",
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

// RequestDeleteUser send empty user token if delete successfully.
func (manager *Manager) RequestDeleteUser(ctx *gin.Context) {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		logger.Error("(RequestDeleteUser)User not login.")
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

// RequestRefreshToken judge whether the refresh token has expired then send fresher token.
func (manager *Manager) RequestRefreshToken(ctx *gin.Context) {
	userId, b := handler.GetUserIdFromTokenIgnoreExpiration(ctx)
	// User id error
	if userId == -1 {
		if b {
			logger.Warn("(RequestRefreshToken)Token error.")
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"code":    http.StatusBadRequest,
					"message": "Token error.",
					"token":   utils.GetUserTokenFromContext(ctx),
				})
		} else {
			logger.Error("(RequestRefreshToken)User not login.")
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
	logger.Trace("Manager.RequestRefreshToken: Refresh token expires at: %v, now: %v",
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

func (manager *Manager) RequestSendVerifyMail(ctx *gin.Context) {
	mailAddr, ok := ctx.GetQuery("mail")
	if !ok || len(mailAddr) == 0 {
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}
	ok = manager.SendVerifyMail(mailAddr)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Send mail failed."})
	} else {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	}
}

func (manager *Manager) RequestGetMailVerify(ctx *gin.Context) {
	var item Item.RequestVerifyMailItem
	err := ctx.ShouldBindJSON(&item)
	if err != nil {
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
		return
	}

	t, code := manager.VerifyMail(item.MailAddr, item.VerifyCode)
	if code == globals.StatusInternalServerError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
		return
	} else if code == globals.StatusNoVerifyCode {
		ctx.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "Mail not found."})
	} else if code == globals.StatusIncorrectVerifyCode {
		ctx.JSON(http.StatusNotAcceptable, gin.H{"code": http.StatusNotAcceptable, "message": "Incorrect verify code."})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "", "mailToken": t})
	}
}
