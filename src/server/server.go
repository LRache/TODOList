package server

import (
	"TODOList/src/globals"
	"TODOList/src/model"
	"TODOList/src/utils"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron/v3"
	"github.com/wonderivan/logger"
	"gopkg.in/gomail.v2"
	"strconv"
	"strings"
	"time"
)

func isUserExists(mailAddr string) bool {
	var userItems []model.DataBaseUserModel
	err := globals.SqlDatabase.Select(&userItems, "SELECT * FROM Users WHERE mailAddr = ? LIMIT 1", mailAddr)
	if err != nil {
		return false
	}
	return len(userItems) != 0
}

func isTodoItemExists(userId int64, itemId int64) bool {
	var count int
	err := globals.SqlDatabase.QueryRow("SELECT COUNT(*) FROM todo WHERE userid = ? AND id = ?", userId, itemId).Scan(&count)
	if err != nil {
		logger.Error("(isTodoItemExists)Error when select from database: %v\n", err.Error())
		return false
	}
	return count != 0
}

// AddItem return new model id and result code
func AddItem(userId int64, todoItem model.RequestTodoItemModel) (int64, int) {
	// Allocate model id
	var newItemId int64
	if globals.RedisClient.LLen(fmt.Sprintf("EmptyItemId:%d", userId)).Val() == 0 {
		newItemId = utils.GetItemCount(userId)
		if newItemId == -1 {
			newItemId = 0
			utils.SetItemCount(userId, 0)
		}
	} else {
		_ = globals.RedisClient.LPop(fmt.Sprintf("EmptyItemId:%d", userId)).Scan(&newItemId)
	}

	// Insert model into database
	logger.Trace("Insert model into database, userId: %v, itemId: %v", userId, newItemId)
	_, err := globals.SqlDatabase.Exec(
		"INSERT INTO todo(id, title, content, create_time, deadline, tag, done, userid) VALUES (?, ?, ?, FROM_UNIXTIME(?), FROM_UNIXTIME(?), ?, ?, ?)",
		newItemId, todoItem.Title, todoItem.Content, todoItem.CreateTime, todoItem.Deadline, todoItem.Tag, todoItem.Done, userId)
	if err != nil {
		logger.Error("Error at insert model:", err.Error())
		return 0, globals.StatusDatabaseCommandError
	}
	utils.SetItemCountPlusOne(userId)
	return newItemId, globals.StatusDatabaseCommandOK
}

// GetItemById return model list and result code.
func GetItemById(userId int64, itemId int64) (model.DataBaseTodoItemModel, int) {
	// Select model from database
	logger.Trace("(GetItemById)Select model from database, userId = %v, itemId = %v", userId, itemId)
	var todoItems []model.DataBaseTodoItemModel
	err := globals.SqlDatabase.Select(&todoItems,
		"SELECT * FROM todo WHERE userId=? AND id=? LIMIT 1", userId, itemId)
	if err != nil {
		logger.Error("Error when select items from database: %v", err.Error())
		return model.DataBaseTodoItemModel{}, globals.StatusDatabaseCommandError
	}

	if len(todoItems) == 0 {
		logger.Warn("model not found: %v\n", itemId)
		return model.DataBaseTodoItemModel{}, globals.StatusItemNotFound
	}
	logger.Trace("(GetItemById)Select model from database successfully, userId = %v, itemId = %v", userId, itemId)
	return todoItems[0], globals.StatusDatabaseCommandOK
}

// GetItems return model list and result code.
func GetItems(userId int64, requestItem model.RequestGetItemsItem, order string, pageIndex int, limit int) ([]model.DataBaseTodoItemModel, int) {
	// Generate select command.
	var command string
	if pageIndex != -1 {
		command = fmt.Sprintf("SELECT * FROM todo WHERE %s ORDER BY %s LIMIT %v, %v",
			strings.Join(append(requestItem.ToSqlSelectWhereCommandStrings(),
				fmt.Sprintf("userid = %d", userId)), " AND "), order, (pageIndex-1)*limit, limit)
	} else {
		command = fmt.Sprintf("SELECT * FROM todo WHERE %s ORDER BY %s",
			strings.Join(append(requestItem.ToSqlSelectWhereCommandStrings(),
				fmt.Sprintf("userid = %d", userId)), " AND "), order)
	}

	// Select items from database.
	logger.Trace("(GetItems)Select items from database, sqlCommand = \"%v\"", command)
	itemList := make([]model.DataBaseTodoItemModel, 0)
	err := globals.SqlDatabase.Select(&itemList, command)
	if err != nil {
		logger.Error("(GetItems)Error when select items from database: %v", err.Error())
		return itemList, globals.StatusDatabaseCommandError
	}

	logger.Trace("(GetItems)Select items from database successfully, userId = %v, count = %v", userId, len(itemList))
	return itemList, globals.StatusDatabaseCommandOK
}

func DeleteItemById(userId int64, itemId int64) int { // Return result code.
	// Ensure model exists
	if !isTodoItemExists(userId, itemId) {
		logger.Warn("(DeleteItemById)model not exists, userId = %v, itemId = %v", userId, itemId)
		return globals.StatusItemNotFound
	}

	// Delete model from database
	logger.Trace("(DeleteItemById)Delete model from database: userId = %v, itemId = %v", userId, itemId)
	_, err := globals.SqlDatabase.Exec("DELETE FROM todo WHERE userid = ? AND id = ?", userId, itemId)
	if err != nil {
		logger.Error("(DeleteItemById)Error when delete model from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	} else {
		logger.Trace("(DeleteItemById)Delete model from database successfully: userId = %v, itemId = %v", userId, itemId)
		// Record empty model id.
		globals.RedisClient.LPush(fmt.Sprintf("EmptyItemId:%d", userId), itemId)
		logger.Trace("Push empty model id to redis: userId = %v, itemId = %v", userId, itemId)
		utils.SetItemCount(userId, utils.GetItemCount(userId)-1)
		return globals.StatusDatabaseCommandOK
	}
}

// UpdateItem return result code.
func UpdateItem(userId int64, itemId int64, values map[string]string) int {
	// Ensure model exists.
	if !isTodoItemExists(userId, itemId) {
		logger.Warn("(UpdateItem)model not exists: userId = %v, itemId = %v", userId, itemId)
		return globals.StatusItemNotFound
	}

	// Update model in database
	// Generate sql command
	command := "UPDATE todo SET "
	valueStrings := make([]string, 0)
	for key, value := range values {
		valueStrings = append(valueStrings, fmt.Sprintf("%s = %s", key, value))
	}
	command += strings.Join(valueStrings, ", ")
	command += fmt.Sprintf(" WHERE userid = %d AND id = %d", userId, itemId)
	logger.Trace("(UpdateItem)Update sql command: \"%s\"", command)

	_, err := globals.SqlDatabase.Exec(command)
	if err != nil {
		logger.Error("(UpdateItem)Error when update database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}
	logger.Trace("(UpdateItem)Update model successfully: userId = %v, itemId = %v", userId, itemId)
	return globals.StatusDatabaseCommandOK
}

// AddUser return new user id, -1 for failure.
func AddUser(user model.RequestRegisterUserItem) int64 {
	var newUserId int64
	// Allocate new user id
	if globals.RedisClient.LLen("EmptyUserId").Val() == 0 {
		newUserId = utils.GetUserCount()
		if newUserId == -1 {
			newUserId = 0
			utils.SetUserCount(0)
		}
	} else {
		_ = globals.RedisClient.LPop("EmptyUserId").Scan(&newUserId)
	}
	logger.Trace("(AddUser)Allocate new userId = %v", newUserId)

	// Insert user into database
	logger.Trace("(AddUser)Insert new user into database: id = %v, name = %v", newUserId, user.Name)
	_, err := globals.SqlDatabase.Exec("INSERT INTO Users(id, username, password, todocount, mailAddr) values(?, ?, ?, 0, ?)",
		newUserId, user.Name, utils.StringToMd5(user.Password), user.MailAddr)
	if err != nil {
		logger.Error("(AddUser)Error when insert user into database: %v", err.Error())
		return -1
	}

	c := utils.SetUserCountPlusOne()
	logger.Trace("(AddUser)Add user successfully, userId = %v, userCount = %v", newUserId, c)
	return newUserId
}

// UserLogin return userid and result code.
func UserLogin(user model.RequestLoginUserItem) (int64, int) {
	// Select from database
	var userItems []model.DataBaseUserModel
	passwordMd5 := utils.StringToMd5(user.Password)

	logger.Trace("(UserLogin)User login: mailAddr = %v, passwordMd5 = %v", user.MailAddr, passwordMd5)
	err := globals.SqlDatabase.Select(&userItems, "SELECT * FROM users WHERE mailAddr = ? AND password = ? LIMIT 1",
		user.MailAddr, passwordMd5)
	if err != nil {
		logger.Error("(UserLogin)Error when select user from database: %v", err.Error())
		return -1, globals.StatusDatabaseCommandError
	}

	if len(userItems) == 0 {
		logger.Warn("Manager.UserLogin: User not found: username = %v", user)
		return -1, globals.StatusItemNotFound
	}
	// Login successfully
	return userItems[0].Id, globals.StatusDatabaseCommandOK
}

func UserReset(mailAddr string, newPassword string) int {
	_, err := globals.SqlDatabase.Exec("UPDATE users SET password = ? WHERE mailAddr = ?",
		utils.StringToMd5(newPassword), mailAddr)
	if err != nil {
		logger.Error("(UserReset)Error when update database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}
	return globals.StatusDatabaseCommandOK
}

// GetUserInfo return user info model and result code.
func GetUserInfo(userId int64) (model.RequestUserInfoItem, int) {
	// Select user from database
	var databaseItems []model.DataBaseUserModel
	var userInfo model.RequestUserInfoItem
	logger.Trace("(GetUserInfo)Select user from database: userId = %v", userId)
	err := globals.SqlDatabase.Select(&databaseItems, "SELECT * FROM users WHERE id = ?", userId)
	if err != nil {
		logger.Error("(GetUserInfo)Error when select from database: %v", err.Error())
		return userInfo, globals.StatusDatabaseCommandError
	}

	if len(databaseItems) == 0 {
		logger.Warn("(GetUserInfo)User not found: userId = %v", userId)
		return userInfo, globals.StatusItemNotFound
	}

	todoCount := utils.GetItemCount(userId)
	if todoCount == -1 {
		todoCount = 0
	}

	databaseItem := databaseItems[0]
	userInfo.UserId = databaseItem.Id
	userInfo.Name = databaseItem.Name
	userInfo.MailAddr = databaseItem.MailAddr
	userInfo.TodoCount = todoCount
	logger.Trace("(GetUserInfo)Load user model successfully: userId = %v", userId)
	return userInfo, globals.StatusDatabaseCommandOK
}

// DeleteUser return result code.
func DeleteUser(userId int64) int {
	var err error

	// Ensure user exists.
	var userItems []model.DataBaseUserModel
	err = globals.SqlDatabase.Select(&userItems, "SELECT * FROM Users WHERE id = ? LIMIT 1", userId)
	if err != nil {
		logger.Error("(DeleteUser)Error when select from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}

	// Delete from database.
	_, err = globals.SqlDatabase.Exec("DELETE FROM users WHERE id = ?", userId)
	if err != nil {
		logger.Error("(DeleteUser)Error when delete user from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}
	_, err = globals.SqlDatabase.Exec("DELETE FROM todo WHERE userid = ?", userId)
	if err != nil {
		logger.Error("(DeleteUser)Error when delete todo items from database: %v", err.Error())
		return globals.StatusDatabaseCommandError
	}

	// Record empty user id.
	globals.RedisClient.LPush("EmptyUserId", userId)
	utils.SetUserCount(utils.GetUserCount() - 1)
	logger.Trace("(DeleteUser)Push empty userid: %v", userId)
	return globals.StatusDatabaseCommandOK
}

// RemoveExpiredVerifyCode scan redis hash "MailVerifyCode" and remove the expired verify code.
func RemoveExpiredVerifyCode() {
	logger.Trace("(RemoveExpiredVerifyCode)Start.")
	now := time.Now().Unix()
	var cursor uint64 = 0
	for {
		values, cursor := globals.RedisClient.HScan("MailVerifyCode", cursor, "", 20).Val()
		for i := 0; i < len(values); i += 2 {
			key := values[i]
			val := values[i+1]
			expiredTime, err := strconv.ParseInt(val[6:], 10, 64)
			if err != nil {
				continue
			}
			if expiredTime < now {
				globals.RedisClient.HDel("MailVerifyCode", key)
				logger.Trace("(RemoveExpiredVerifyCode)Remove verify code, mailAddr = \"%s\"", key)
			}
		}
		if cursor == 0 {
			break
		}
	}
}

// SendVerifyMail email addr with verify code, return if send successfully.
func SendVerifyMail(addr string) bool {
	code := utils.GenerateRandomVerifyCode()

	m := gomail.NewMessage()
	m.SetHeader("From", globals.MailFrom)
	m.SetHeader("To", addr)
	m.SetHeader("Subject", "Verify Your Email")
	m.SetBody("text/html", fmt.Sprintf("You verify code is <br> <b>%s</b>", code))
	logger.Trace("(SendVerifyMail)Send verify mail to \"%s\", verifyCode = \"%s\"", addr, code)
	err := gomail.Send(*globals.MailSender, m)
	if err != nil {
		logger.Warn("(SendVerifyMail)Error when send mail: %v", err.Error())
		return false
	}

	expiredTime := time.Now().Add(globals.MailVerifyCodeValidity).Unix()
	expiredTimeString := fmt.Sprintf("%d", expiredTime)
	ok := globals.RedisClient.HSet("MailVerifyCode", addr, code+expiredTimeString).Val()
	if ok {
		logger.Trace("(SendVerifyMail)Push verify code successfully.")
	} else {
		logger.Error("(SendVerifyMail)Error when push verify code.")
	}

	return ok
}

// VerifyMail verify mail code, return mail token and if code is correct.
func VerifyMail(mailAddr string, code string) (string, int) {
	// Get verify code from database
	cmd := globals.RedisClient.HGet("MailVerifyCode", mailAddr)
	if cmd.Err() != nil {
		return "", globals.StatusNoVerifyCode
	}

	value := cmd.Val()
	correctCode := value[:6]
	if correctCode == "" {
		logger.Trace("(VerifyMail)Mail not found: \"%v\"", mailAddr)
		return "", globals.StatusNoVerifyCode
	}
	if correctCode != code {
		logger.Trace("(VerifyMail)Incorrect verify code: \"%v\", correct code is \"%v\"", code, correctCode)
		return "", globals.StatusIncorrectVerifyCode
	}
	t, ok := GenerateMailVerifyCodeToken(&MailVerifyCodeClaims{MailAddr: mailAddr})
	if !ok {
		logger.Error("(VerifyMail)Error when generate token.")
		return "", globals.StatusInternalServerError
	}
	globals.RedisClient.HDel("MailVerifyCode", mailAddr)
	logger.Trace("(VerifyMail)Verified mail: \"%v\".", mailAddr)
	return t, globals.StatusOK
}

// SetItemCron Set a schedule which will call itemCronFun at model deadline before a duration
func SetItemCron(userId int64, itemId int64, d time.Duration) int {
	todoItem, code := GetItemById(userId, itemId)
	if code != globals.StatusDatabaseCommandOK {
		return code
	}
	deadlineTime, _ := time.Parse(time.DateTime, todoItem.Deadline)

	deadlineTime = deadlineTime.Add(-d)
	month := deadlineTime.Month()
	day := deadlineTime.Day()
	hour := deadlineTime.Hour()
	minute := deadlineTime.Minute()
	second := deadlineTime.Second()
	c := cron.New(cron.WithSeconds())
	err := utils.GenerateOnceCron(c, fmt.Sprintf("%d %d %d %d %d ?", second, minute, hour, day, month), func() {
		itemCronFun(userId, todoItem)
	})
	if err != nil {
		logger.Error("(SetItemCron)Error when set cron: %v", err.Error())
		return globals.StatusInternalServerError
	}
	c.Start()
	logger.Trace("(SetItemCron)Set model cron successfully, userId = %v, itemId = %v", userId, itemId)

	return globals.StatusOK
}

// itemCronFun Do sth.
func itemCronFun(userId int64, todoItem model.DataBaseTodoItemModel) {
	logger.Trace("(itemCronFun)Item cron, userId = %d, todoItem = %v", userId, todoItem)
	/* Do sth,
	like email the user or put some information to the redis task queue, so that another program can send message to app
	or remind user when he or she logs in on the website.
	*/

}
