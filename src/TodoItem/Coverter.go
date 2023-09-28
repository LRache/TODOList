package TodoItem

func DatabaseToRequestTodoItem(databaseItem DataBaseTodoItem) RequestTodoItem {
	var requestItem RequestTodoItem
	requestItem.Id = databaseItem.Id
	requestItem.Title = databaseItem.Title
	requestItem.Content = databaseItem.Content
	requestItem.CreateTime = databaseItem.CreateTime
	requestItem.Deadline = databaseItem.Deadline
	requestItem.Tag = databaseItem.Tag
	requestItem.Done = databaseItem.Done
	return requestItem
}

func ListDatabaseToRequestTodoItem(databaseItems []DataBaseTodoItem) []RequestTodoItem {
	requestItems := make([]RequestTodoItem, len(databaseItems))
	for index, databaseItem := range databaseItems {
		requestItems[index] = DatabaseToRequestTodoItem(databaseItem)
	}
	return requestItems
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
