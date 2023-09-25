package Todo

import (
	"fmt"
)

type RequestTodoItem struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreateTime string `json:"createTime"`
	Deadline   string `json:"deadline"`
	Tag        string `json:"tag"`
	Done       bool   `json:"done"`
}

func (requestItem RequestTodoItem) Output() {
	fmt.Println("title: ", requestItem.Title)
	fmt.Println("content: ", requestItem.Content)
	fmt.Println("createTime", requestItem.CreateTime)
	fmt.Println("deadline", requestItem.Deadline)
	fmt.Println("tag", requestItem.Tag)
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
