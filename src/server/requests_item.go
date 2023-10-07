package server

import (
	"TODOList/src/globals"
	"TODOList/src/model"
	"TODOList/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"net/http"
	"strconv"
	"time"
)

// checkUserLogin return -1 if user not login, and set context.
func checkUserLogin(ctx *gin.Context) int64 {
	userId := utils.GetUserIdFromContext(ctx)
	if userId == -1 {
		logger.Info("User not login.")
		ctx.JSON(globals.ReturnJsonUserNotLogin.Code, globals.ReturnJsonUserNotLogin.Json)
		return -1
	}
	return userId
}

// RequestAddItem send new model id.
func RequestAddItem(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	var todoItem model.RequestTodoItemModel
	var err error
	err = ctx.ShouldBindJSON(&todoItem)
	if err != nil {
		logger.Warn("(RequestAddItem)Bind body json error: %v", err.Error())
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
		return
	}

	itemId, code := AddItem(userId, todoItem)
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

// RequestGetItemById send a model using RequestTodoItemModel type.
func RequestGetItemById(ctx *gin.Context) {
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

	todoDatabaseItem, code := GetItemById(userId, itemId)
	if code == globals.StatusDatabaseCommandOK {
		requestItem := todoDatabaseItem.ToRequestTodoModel()
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "model": requestItem})
	} else if code == globals.StatusItemNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

// RequestGetItems send model list using RequestTodoItemModel type.
func RequestGetItems(ctx *gin.Context) {
	userid := checkUserLogin(ctx)
	if userid == -1 {
		return
	}

	var requestItem model.RequestGetItemsModel
	err := ctx.ShouldBindQuery(&requestItem)
	if err != nil {
		logger.Warn("(RequestGetItems)Error when bind query: %v", err.Error())
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}

	// set page index and limit
	var pageIndex, limit int
	if requestItem.PageIndex == "" {
		pageIndex = -1
	} else {
		if pageIndex, err = strconv.Atoi(requestItem.PageIndex); err != nil {
			ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
			return
		}
		if requestItem.Limit == "" {
			ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
			return
		} else {
			if limit, err = strconv.Atoi(requestItem.Limit); err != nil {
				ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
				return
			}
		}
	}

	// check order
	var order string
	if requestItem.Order == "" {
		order = "id"
	} else if requestItem.Order != "id" && requestItem.Order != "deadline" && requestItem.Order != "createTime" {
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}

	items, code := GetItems(userid, requestItem, order, pageIndex, limit)
	if code == globals.StatusDatabaseCommandError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else {
		ctx.JSON(http.StatusOK,
			gin.H{
				"code":    http.StatusOK,
				"message": "",
				"items":   model.ListDatabaseToRequestTodoItem(items),
			})
	}
}

// RequestUpdateItem send code and message.
func RequestUpdateItem(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	// Parse body
	var requestItem model.RequestUpdateTodoItemModel
	err := ctx.ShouldBindJSON(&requestItem)
	if err != nil {
		logger.Warn("(RequestUpdateItem)Error when bind body json: %v", err.Error())
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
	}

	// Select items from database
	code := UpdateItem(userId, requestItem.ItemId, requestItem.ToDatabaseMap())
	if code == globals.StatusItemNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else if code == globals.StatusDatabaseCommandError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	}
}

// RequestDeleteItemById send code and message.
func RequestDeleteItemById(ctx *gin.Context) {
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

	// Delete model from database
	code := DeleteItemById(userId, itemId)
	if code == globals.StatusDatabaseCommandOK {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	} else if code == globals.StatusItemNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

func RequestSetItemCron(ctx *gin.Context) {
	userId := checkUserLogin(ctx)
	if userId == -1 {
		return
	}

	itemIdString, ok := ctx.GetQuery("id")
	if !ok {
		logger.Warn("(RequestSetItemCron)Lack of query itemId.")
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}
	itemId, err := strconv.ParseInt(itemIdString, 10, 64)
	if err != nil {
		logger.Warn("(RequestSetItemCron)Error when parse param: %v", err.Error())
		ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
		return
	}

	var before time.Duration
	beforeString, ok := ctx.GetQuery("before")
	if ok {
		before, err = time.ParseDuration(beforeString)
		if err != nil {
			logger.Warn("(RequestSetItemCron)Error when parse param(before duration): %v", err.Error())
			ctx.JSON(globals.ReturnJsonQueryError.Code, globals.ReturnJsonQueryError.Json)
			return
		}
	} else {
		before = time.Duration(0)
	}

	code := SetItemCron(userId, itemId, before)
	if code == globals.StatusItemNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else if code == globals.StatusInternalServerError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else {
		ctx.JSON(http.StatusCreated, globals.ReturnJsonSuccess.Json)
	}
}
