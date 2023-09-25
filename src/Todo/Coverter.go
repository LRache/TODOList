package Todo

func DatabaseToRequestTodoItem(databaseItem DataBaseTodoItem) RequestTodoItem {
	var requestItem RequestTodoItem
	requestItem.Title = databaseItem.Title
	requestItem.Content = databaseItem.Content
	requestItem.CreateTime = databaseItem.CreateTime
	requestItem.Deadline = databaseItem.Deadline
	requestItem.Tag = databaseItem.Tag
	requestItem.Done = databaseItem.Done
	return requestItem
}

func RequestToTodoItem(requestItem RequestTodoItem) Item {
	var item Item
	item.Title = requestItem.Title
	item.Content = requestItem.Content
	item.CreateTime = requestItem.CreateTime
	item.Deadline = requestItem.Deadline
	item.Tag = requestItem.Tag
	item.Done = requestItem.Done
	return item
}
