package Todo

import (
	"fmt"
	"strings"
)

type RequestTodoItem struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreateTime string `json:"createTime"`
	Deadline   string `json:"deadline"`
	Tag        string `json:"tag"`
	Done       bool   `json:"done"`
}

func (requestItem *RequestTodoItem) Output() {
	fmt.Println("title: ", requestItem.Title)
	fmt.Println("content: ", requestItem.Content)
	fmt.Println("createTime: ", requestItem.CreateTime)
	fmt.Println("deadline: ", requestItem.Deadline)
	fmt.Println("tag: ", requestItem.Tag)
	fmt.Println("done: ", requestItem.Done)
}

type RequestUpdateTodoItem struct {
	UpdateKeys []string `json:"updateKeys"`
	ItemId     int      `json:"itemId"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	CreateTime string   `json:"createTime"`
	Deadline   string   `json:"deadline"`
	Tag        string   `json:"tag"`
	Done       bool     `json:"done"`
}

func (requestItem *RequestUpdateTodoItem) Output() {
	fmt.Println("updateKeys: ", requestItem.UpdateKeys)
	fmt.Println("title: ", requestItem.Title)
	fmt.Println("content: ", requestItem.Content)
	fmt.Println("createTime: ", requestItem.CreateTime)
	fmt.Println("deadline: ", requestItem.Deadline)
	fmt.Println("tag: ", requestItem.Tag)
	fmt.Println("done: ", requestItem.Done)
}

func (requestItem *RequestUpdateTodoItem) ToDataBaseMap() map[string]string {
	m := make(map[string]string)
	for _, key := range requestItem.UpdateKeys {
		switch key {
		case "title":
			m["title"] = fmt.Sprintf("\"%s\"", requestItem.Title)
		case "content":
			m["content"] = fmt.Sprintf("\"%s\"", requestItem.Content)
		case "createTime":
			m["createTime"] = fmt.Sprintf("\"%s\"", requestItem.CreateTime)
		case "deadline":
			m["deadline"] = fmt.Sprintf("\"%s\"", requestItem.Deadline)
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

func (requestItem *RequestUpdateTodoItem) ToSqlCommandString() string {
	commandStrings := make([]string, 0)
	for _, key := range requestItem.UpdateKeys {
		switch key {
		case "title":
			commandStrings = append(commandStrings, fmt.Sprintf("title = \"%s\"", requestItem.Title))
		case "content":
			commandStrings = append(commandStrings, fmt.Sprintf("content = \"%s\"", requestItem.Content))
		case "createTime":
			commandStrings = append(commandStrings, fmt.Sprintf("createTime = \"%s\"", requestItem.CreateTime))
		case "deadline":
			commandStrings = append(commandStrings, fmt.Sprintf("deadline = \"%s\"", requestItem.Deadline))
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

type RequestUserNameItem struct {
	Name string `json:"username"`
}

type RequestLoginUserItem struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}

type RequestUserInfoItem struct {
	Name      string `json:"username"`
	UserId    int    `json:"userid"`
	TodoCount int    `json:"todoCount"`
}

type RequestGetItemsItem struct {
	Tag  string `query:"tag"`
	Done string `query:"done"`
}

func (requestItem *RequestGetItemsItem) ToSqlSelectWhereCommand() string {
	commandStrings := make([]string, 0)
	if requestItem.Tag != "" {
		commandStrings = append(commandStrings, fmt.Sprintf("tag = \"%s\"", requestItem.Tag))
	}
	if requestItem.Done != "" {
		commandStrings = append(commandStrings, fmt.Sprintf("done = \"%v\""), requestItem.Done)
	}
	if len(commandStrings) == 0 {
		return ""
	} else {
		return "WHERE " + strings.Join(commandStrings, " AND ")
	}
}
