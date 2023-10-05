package server

import (
	"TODOList/src/globals"
	"TODOList/src/handler"
	"TODOList/src/model"
	"TODOList/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"net/http"
	"strings"
	"time"
)

// RequestRegisterUser send user token in json and fresh refreshToken
func RequestRegisterUser(ctx *gin.Context) {
	var userItem model.RequestRegisterUserItem
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
	if isUserExists(userItem.MailAddr) {
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
		mailFromToken, ok := GetMailAddrFromToken(userItem.MailToken)
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
		} else if mailFromToken != userItem.MailAddr {
			logger.Trace("(RequestRegisterUser)Unmatched token, mailInToken = %v but got %v",
				mailFromToken, userItem.MailAddr)
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
	newUserId := AddUser(userItem)
	if newUserId != -1 {
		ctx.JSON(http.StatusCreated,
			gin.H{
				"code":         http.StatusCreated,
				"userId":       newUserId,
				"token":        utils.GenerateUserToken(newUserId),
				"refreshToken": utils.GenerateUserRefreshToken(newUserId),
			})
		logger.Trace("(RequestRegisterUser)Add user successfully: %v", userItem.Name)
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
func RequestLogin(ctx *gin.Context) {
	var userItem model.RequestLoginUserItem
	err := ctx.ShouldBindJSON(&userItem)
	if err != nil {
		logger.Warn("(RequestLogin)Error when bind json: %v", err.Error())
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
	userId, code := UserLogin(userItem)
	if code == globals.StatusDatabaseCommandOK { // Login successfully
		logger.Trace("(RequestLogin)User login successfully: %v", userItem.MailAddr)
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
		return
	}
	if code == globals.StatusItemNotFound {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":         http.StatusUnauthorized,
				"message":      "Incorrect username or password.",
				"token":        utils.GetUserTokenFromContext(ctx),
				"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
			})
		return
	}
	ctx.JSON(
		http.StatusInternalServerError,
		gin.H{
			"code":         http.StatusInternalServerError,
			"message":      "Internal server error.",
			"token":        utils.GetUserTokenFromContext(ctx),
			"refreshToken": utils.GetUserRefreshTokenStringFromContext(ctx),
		})

}

func RequestGetCurrentUser(ctx *gin.Context) {
	userId := utils.GetUserIdFromContext(ctx)

	// User not login, return userId=-1 means no user
	if userId == -1 {
		logger.Info("(RequestGetCurrentUser)User not login.")
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
	userInfo, code := GetUserInfo(userId)
	if code == globals.StatusItemNotFound {
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
				"userinfo": userInfo,
			})
	}
}

func RequestResetUser(ctx *gin.Context) {
	var requestItem model.RequestResetUserItem
	if err := ctx.ShouldBindJSON(&requestItem); err != nil {
		logger.Warn("(RequestResetUser)Error when parse body json: %v", err.Error())
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
		return
	}

	if !isUserExists(requestItem.MailAddr) {
		logger.Info("(RequestResetUser)User not found: \"%v\"", requestItem.MailAddr)
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound)
		return
	}
	if !strings.HasSuffix(requestItem.MailAddr, "@todouser") {
		mailFromToken, ok := GetMailAddrFromToken(requestItem.MailToken)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Parse token error."})
			return
		}
		if mailFromToken != requestItem.MailAddr {
			logger.Trace("(RequestRegisterUser)Unmatched token, mailInToken = %v but got %v",
				mailFromToken, requestItem.MailAddr)
			ctx.JSON(http.StatusNotAcceptable, gin.H{"code": http.StatusNotAcceptable, "message": "Unmatched token and mail address."})
			return
		}
	}

	code := UserReset(requestItem.MailAddr, requestItem.NewPassword)
	if code == globals.StatusDatabaseCommandOK {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

// RequestDeleteUser send empty user token if delete successfully.
func RequestDeleteUser(ctx *gin.Context) {
	userId := utils.GetUserIdFromContext(ctx)
	if userId == -1 {
		logger.Info("(RequestDeleteUser)User not login.")
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Not login.",
				"token":   utils.GenerateEmptyUserToken(),
			})
		return
	}

	code := DeleteUser(userId)
	if code != globals.StatusDatabaseCommandOK {

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Internal server error.",
				"token":   utils.GetUserTokenFromContext(ctx),
			})
	} else {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "",
				"token":   utils.GenerateEmptyUserToken(),
			})
	}
}

// RequestRefreshToken judge whether the refresh token has expired then send fresher token.
func RequestRefreshToken(ctx *gin.Context) {
	userId, b := utils.GetUserIdFromContextIgnoreExpiration(ctx)
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
			logger.Info("(RequestRefreshToken)User not login.")
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

	// Refresh token has unmatched user id and token code
	if !utils.CheckUserTokenCodeString(c.UserId, c.UserTokenCode) {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Refresh token has unmatched user id and token code.",
				"token":   utils.GetUserTokenFromContext(ctx),
			})
		return
	}

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

	// Refresh token and user token unmatched.
	if c.UserId != userId {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid refresh token.",
				"token":   utils.GetUserTokenFromContext(ctx),
			})
		return
	}

	// Refresh successfully.
	ctx.JSON(
		http.StatusOK,
		gin.H{
			"code":    http.StatusOK,
			"message": "Refresh token successfully.",
			"token":   utils.GenerateUserToken(userId),
		})
}

func RequestSendVerifyMail(ctx *gin.Context) {
	mailAddr, ok := ctx.GetQuery("mail")

	// Invalid mail address.
	if !ok || len(mailAddr) == 0 || !utils.IsMailFormat(mailAddr) {
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}

	ok = SendVerifyMail(mailAddr)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Send mail failed."})
	} else {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	}
}

func RequestGetMailVerify(ctx *gin.Context) {
	var verifyMailItem model.RequestVerifyMailItem
	err := ctx.ShouldBindJSON(&verifyMailItem)
	if err != nil {
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
		return
	}

	t, code := VerifyMail(verifyMailItem.MailAddr, verifyMailItem.VerifyCode)
	if code == globals.StatusInternalServerError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else if code == globals.StatusNoVerifyCode {
		ctx.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "Mail not found."})
	} else if code == globals.StatusIncorrectVerifyCode {
		ctx.JSON(http.StatusNotAcceptable, gin.H{"code": http.StatusNotAcceptable, "message": "Incorrect verify code."})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "", "mailToken": t})
	}
}
