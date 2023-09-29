package server

import (
	"TODOList/src/Item"
	"TODOList/src/globals"
	"TODOList/src/handler"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"net/http"
	"strconv"
)

// checkUserLogin return -1 if user not login, and set context.
func checkUserLogin(ctx *gin.Context) int64 {
	userId := handler.GetUserIdFromToken(ctx)
	if userId == -1 {
		logger.Info("User not login.")
		ctx.JSON(globals.ReturnJsonUserNotLogin.Code, globals.ReturnJsonUserNotLogin.Json)
		return -1
	}
	return userId
}

// RequestAddItem send new item id.
func RequestAddItem(ctx *gin.Context) {
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

	itemId, code := AddItem(userId, Item.RequestToTodoItem(item))
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
		requestItem := Item.DatabaseToRequestTodoItem(todoDatabaseItem)
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "item": requestItem})
	} else if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}

// RequestGetItems send item list using RequestTodoItem type.
func RequestGetItems(ctx *gin.Context) {
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
				"items":   Item.ListDatabaseToRequestTodoItem(items),
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
	var requestItem Item.RequestUpdateTodoItem
	err := ctx.ShouldBindJSON(&requestItem)
	if err != nil {
		logger.Warn("(RequestUpdateItem)Error when bind body json: %v", err.Error())
		ctx.JSON(globals.ReturnJsonBodyJsonError.Code, globals.ReturnJsonBodyJsonError.Json)
	}

	// Select items from database
	code := UpdateItem(userId, requestItem.ItemId, requestItem.ToDataBaseMap())
	if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else if code == globals.StatusDatabaseCommandError {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	} else {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	}
}

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

	// Delete item from database
	code := DeleteItemById(userId, itemId)
	if code == globals.StatusDatabaseCommandOK {
		ctx.JSON(globals.ReturnJsonSuccess.Code, globals.ReturnJsonSuccess.Json)
	} else if code == globals.StatusDatabaseSelectNotFound {
		ctx.JSON(globals.ReturnJsonItemNotFound.Code, globals.ReturnJsonItemNotFound.Json)
	} else {
		ctx.JSON(globals.ReturnJsonInternalServerError.Code, globals.ReturnJsonInternalServerError.Json)
	}
}
