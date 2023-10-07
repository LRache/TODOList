package model

import (
	"time"
)

func (databaseItem *DataBaseTodoItemModel) ToRequestTodoModel() RequestTodoItemModel {
	createTime, _ := time.ParseInLocation(time.DateTime, databaseItem.CreateTime, time.Local)
	deadline, _ := time.ParseInLocation(time.DateTime, databaseItem.Deadline, time.Local)
	var requestItem RequestTodoItemModel
	requestItem.Id = databaseItem.Id
	requestItem.Title = databaseItem.Title
	requestItem.Content = databaseItem.Content
	requestItem.CreateTime = createTime.Unix()
	requestItem.Deadline = deadline.Unix()
	requestItem.Tag = databaseItem.Tag
	requestItem.Done = databaseItem.Done
	return requestItem
}

func (requestItem *RequestTodoItemModel) ToDatabaseTodoModel() DataBaseTodoItemModel {
	var databaseItem DataBaseTodoItemModel
	databaseItem.Id = requestItem.Id
	databaseItem.Title = requestItem.Title
	databaseItem.Content = requestItem.Content
	databaseItem.CreateTime = time.Unix(requestItem.CreateTime, 0).Format(time.DateTime)
	databaseItem.Deadline = time.Unix(requestItem.Deadline, 0).Format(time.DateTime)
	databaseItem.Tag = requestItem.Tag
	databaseItem.Done = requestItem.Done
	return databaseItem
}

func ListDatabaseToRequestTodoItem(databaseItems []DataBaseTodoItemModel) []RequestTodoItemModel {
	requestItems := make([]RequestTodoItemModel, len(databaseItems))
	for index, databaseItem := range databaseItems {
		requestItems[index] = databaseItem.ToRequestTodoModel()
	}
	return requestItems
}
