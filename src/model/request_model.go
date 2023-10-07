package model

import (
	"fmt"
	"strings"
)

type RequestTodoItemModel struct {
	Id         int64  `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreateTime int64  `json:"createTime"`
	Deadline   int64  `json:"deadline"`
	Tag        string `json:"tag"`
	Done       bool   `json:"done"`
}

func (requestItem *RequestTodoItemModel) Output() {
	fmt.Println("title: ", requestItem.Title)
	fmt.Println("content: ", requestItem.Content)
	fmt.Println("createTime: ", requestItem.CreateTime)
	fmt.Println("deadline: ", requestItem.Deadline)
	fmt.Println("tag: ", requestItem.Tag)
	fmt.Println("done: ", requestItem.Done)
}

type RequestUpdateTodoItemModel struct {
	UpdateKeys []string `json:"updateKeys"`
	ItemId     int64    `json:"itemId"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	CreateTime int64    `json:"createTime"`
	Deadline   int64    `json:"deadline"`
	Tag        string   `json:"tag"`
	Done       bool     `json:"done"`
}

func (requestItem *RequestUpdateTodoItemModel) Output() {
	fmt.Println("updateKeys: ", requestItem.UpdateKeys)
	fmt.Println("title: ", requestItem.Title)
	fmt.Println("content: ", requestItem.Content)
	fmt.Println("createTime: ", requestItem.CreateTime)
	fmt.Println("deadline: ", requestItem.Deadline)
	fmt.Println("tag: ", requestItem.Tag)
	fmt.Println("done: ", requestItem.Done)
}

func (requestItem *RequestUpdateTodoItemModel) ToDatabaseMap() map[string]string {
	m := make(map[string]string)
	for _, key := range requestItem.UpdateKeys {
		switch key {
		case "title":
			m["title"] = fmt.Sprintf("\"%s\"", requestItem.Title)
		case "content":
			m["content"] = fmt.Sprintf("\"%s\"", requestItem.Content)
		case "createTime":
			m["createTime"] = fmt.Sprintf("FROM_UNIXTIME(%d)", requestItem.CreateTime)
		case "deadline":
			m["deadline"] = fmt.Sprintf("FROM_UNIXTIME(%d)", requestItem.Deadline)
		case "tag":
			m["tag"] = fmt.Sprintf("\"%s\"", requestItem.Tag)
		case "done":
			if requestItem.Done {
				m["done"] = "1"
			} else {
				m["done"] = "0"
			}
		}
	}
	return m
}

func (requestItem *RequestUpdateTodoItemModel) ToSqlCommandString() string {
	commandStrings := make([]string, 0)
	for _, key := range requestItem.UpdateKeys {
		switch key {
		case "title":
			commandStrings = append(commandStrings, fmt.Sprintf("title = \"%s\"", requestItem.Title))
		case "content":
			commandStrings = append(commandStrings, fmt.Sprintf("content = \"%s\"", requestItem.Content))
		case "createTime":
			commandStrings = append(commandStrings, fmt.Sprintf("createTime = \"%d\"", requestItem.CreateTime))
		case "deadline":
			commandStrings = append(commandStrings, fmt.Sprintf("deadline = \"%d\"", requestItem.Deadline))
		case "tag":
			commandStrings = append(commandStrings, fmt.Sprintf("tag = \"%s\"", requestItem.Tag))
		case "done":
			if requestItem.Done {
				commandStrings = append(commandStrings, fmt.Sprintf("done = 1"))
			} else {
				commandStrings = append(commandStrings, fmt.Sprintf("done = 0"))
			}
		}
	}
	return strings.Join(commandStrings, ", ")
}

type RequestUserNameModel struct {
	Name string `json:"username"`
}

type RequestRegisterUserModel struct {
	Name      string `json:"username"`
	Password  string `json:"password"`
	MailAddr  string `json:"mailAddr"`
	MailToken string `json:"mailToken"`
}

type RequestLoginUserModel struct {
	MailAddr string `json:"mailAddr"`
	Password string `json:"password"`
}

type RequestUserInfoModel struct {
	Name      string `json:"username"`
	UserId    int64  `json:"userid"`
	TodoCount int64  `json:"todoCount"`
	MailAddr  string `json:"mailAddr"`
}

type RequestGetItemsModel struct {
	Tag       string `form:"tag"`
	Done      string `form:"done"`
	Deadline  string `form:"deadlineBefore"`
	PageIndex string `form:"pageIndex"`
	Limit     string `form:"limit"`
	Order     string `form:"order"`
}

func (requestItem *RequestGetItemsModel) ToSqlSelectWhereCommandStrings() []string {
	commandStrings := make([]string, 0)
	if requestItem.Tag != "" {
		commandStrings = append(commandStrings, fmt.Sprintf("tag = \"%s\"", requestItem.Tag))
	}
	if requestItem.Deadline != "" {
		commandStrings = append(commandStrings, fmt.Sprintf("deadline < \"%s\"", requestItem.Deadline))
	}
	if requestItem.Done != "" {
		if requestItem.Done == "true" {
			commandStrings = append(commandStrings, "done = 1")
		} else if requestItem.Done == "false" {
			commandStrings = append(commandStrings, "done = 0")
		}
	}
	return commandStrings
}

type RequestVerifyMailItemModel struct {
	MailAddr   string `json:"mail"`
	VerifyCode string `json:"code"`
}

type RequestResetUserItemModel struct {
	MailAddr    string `json:"mailAddr"`
	MailToken   string `json:"mailToken"`
	NewPassword string `json:"newPassword"`
}
