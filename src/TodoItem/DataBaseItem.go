package TodoItem

type DataBaseUserItem struct {
	Id        int64  `db:"id"`
	Name      string `db:"username"`
	Password  string `db:"password"`
	TodoCount int64  `db:"todocount"`
}

type DataBaseTodoItem struct {
	Title      string `db:"title"`
	Content    string `db:"content"`
	CreateTime string `db:"create_time"`
	Deadline   string `db:"deadline"`
	Tag        string `db:"tag"`
	KeyId      int64  `db:"keyid"`
	Id         int64  `db:"id"`
	UserId     int64  `db:"userid"`
	Done       bool   `db:"done"`
}
