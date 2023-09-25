package Todo

type DataBaseUserItem struct {
	Id       int    `db:"id"`
	Name     string `db:"username"`
	Password string `db:"password"`
}

type DataBaseTodoItem struct {
	Title      string `db:"title"`
	Content    string `db:"content"`
	CreateTime string `db:"create_time"`
	Deadline   string `db:"deadline"`
	Tag        string `db:"tag"`
	KeyId      int    `db:"keyid"`
	Id         int    `db:"id"`
	UserId     int    `db:"userid"`
	Done       bool   `db:"done"`
}
