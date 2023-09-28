# TODOList
## 数据结构
### TODOItem

|     字段     |   类型   |   描述    |
|:----------:|:------:|:-------:|
|     id     |  int   | TODO id |
|   title    | string |   标题    |
|  content   |  int   |   内容    |
| createTime | string |  创建时间   |
|  deadline  | string |  截止时间   |
|    tag     | string |   标签    |
|    done    |  bool  |  是否完成   |
## 接口API
### 注册
路由：`/todo/user`  
方法：`PUT`  
参数：

|    位置     |    字段    |   类型   |  描述  |
|:---------:|:--------:|:------:|:----:|
| Body Json | username | string | 用户名称 |
| Body Json | password | string | 用户密码 |
返回：

|      字段      |   类型   |         描述          |
|:------------:|:------:|:-------------------:|
|     code     |  int   |       成功则为201       |
|   message    | string |                     |
|    userId    |  int   |       新用户的id        |
|    token     | string |    用于自动登录的token     |
| refreshToken | string | 用于在token过期后刷新的token |

### 登录和注销
路由： `/todo/user`  
方法：`POST`  
参数：  
Body Json

|    位置     |    字段    |   类型   |        描述        |
|:---------:|:--------:|:------:|:----------------:|
| Body Json | username | string | 用户名称, 如果为空，则退出登录 |
| Body Json | password | string |       用户密码       |  

返回：

|      字段      |   类型   |         描述          |
|:------------:|:------:|:-------------------:|
|     code     |  int   |       成功则为200       |
|   message    | string |                     |
|    userId    |  int   |       登录用户的id       |
|    token     | string |    用于自动登录的token     |
| refreshToken | string | 用于在token过期后刷新的token |     

### 获取登录用户信息
路由：`/todo/user`  
方法：`GET`  
参数：

|   位置    |  字段   |   类型   |   描述    |
|:-------:|:-----:|:------:|:-------:|
| Headers | token | string | 用户token |
返回：

|         字段         |   类型   |         描述          |
|:------------------:|:------:|:-------------------:|
|        code        |  int   |       成功则为200       |
|      message       | string |                     |
|  userinfo.userId   |  int   |        用户的id        |
| userinfo.username  | string |    用于自动登录的token     |
| userinfo.todoCount | string | 用于在token过期后刷新的token |    

### 删除用户
路由：`/todo/user`  
方法：`DELETE`
参数：

|   位置    |  字段   |   类型   |   描述    |
|:-------:|:-----:|:------:|:-------:|
| Headers | token | string | 用户token |

返回：

|      字段      |   类型   |              描述              |
|:------------:|:------:|:----------------------------:|
|     code     |  int   |           成功则为200            |
|   message    | string |                              |
|    token     | string | 用于自动登录的token,删除成功则为空的用户token | 

### 更新用户token
路由：`/todo/user/token`  
方法：`POST`
参数：

|   位置    |      字段      |   类型   |           描述           |
|:-------:|:------------:|:------:|:----------------------:|
| Headers |    token     | string |        用户token         |
| Headers | refreshToken | string | 用于刷新token的refreshToken |

返回：

|      字段      |   类型   |    描述     |
|:------------:|:------:|:---------:|
|     code     |  int   |  成功则为200  |
|   message    | string |           |
|    token     | string | 更新后的token |

### 添加TODO
路由：`/todo/item`  
方法：`PUT`
参数：

|    位置     |     字段     |   类型   |    描述    |
|:---------:|:----------:|:------:|:--------:|
|  Headers  |   token    | string | 用户token  |
| Body Json |   title    | string |  TODO标题  |
| Body Json |  content   | string |  TODO内容  | 
| Body Json | createTime | string | TODO创建时间 |
| Body Json |  deadline  | string | TODO截止时间 |
| Body Json |    done    |  bool  | TODO是否完成 |

返回：

|   字段    |   类型   |    描述     |
|:-------:|:------:|:---------:|
|  code   |  int   |  成功则为201  |
| message | string |           |
| userId  |  int   | 操作的用户的id  |
| itemId  |  int   | 新的TODO的id |

### 通过id获取TODO
路由：`/todo/item/{:id}`    
方法：`GET`  
参数：

|    位置     |     字段     |   类型   |    描述    |
|:---------:|:----------:|:------:|:--------:|
|  Headers  |   token    | string | 用户token  |
|   Path    |     id     |  int   | TODO id  |
返回：

|   字段    |   类型   |     描述      |
|:-------:|:------:|:-----------:|
|  code   |  int   |   成功则为200   |
| message | string |             |
|  item   | object | TODOItem 类型 |

### 获取TODO（可筛选）
路由：`/todo/item`  
方法：`GET`  
参数：

|    位置     |     字段     |   类型   |     描述     |
|:---------:|:----------:|:------:|:----------:|
|  Headers  |   token    | string |  用户token   |
|   Query   |    tag     | string |  留空则不筛选标题  |
|   Query   |    done    |  bool  | 留空则不筛选是否完成 |
返回：

|   字段    |       类型       |      描述       |
|:-------:|:--------------:|:-------------:|
|  code   |      int       |    成功则为200    |
| message |     string     |               |
|  items  | list[TODOItem] | TODOItem 类型列表 |

### 更新TODO
路由：`/todo/item`  
方法：`POST`  
参数：

|    位置     |     字段     |      类型      |     描述     |
|:---------:|:----------:|:------------:|:----------:|
|  Headers  |   token    |    string    |  用户token   |
| Body Json |   itemId   |     int      |  TODO id   |
| Body Json | updateKeys | list[string] | 要更新字段的名称列表 |
| Body Json |   title    |    string    |   TODO标题   |
| Body Json |  content   |    string    |   TODO内容   |
| Body Json | createTime |    string    |  TODO创建时间  |
| Body Json |  deadline  |    string    |  TODO截止时间  |
| Body Json |    done    |     bool     |  TODO是否完成  |
返回：

|      字段      |   类型   |    描述     |
|:------------:|:------:|:---------:|
|     code     |  int   |  成功则为200  |
|   message    | string |           |

### 删除TODO
路由：`/todo/item/{:id}`  
方法：`DELETE`  
参数：

|   位置    |  字段   |   类型   |   描述    |
|:-------:|:-----:|:------:|:-------:|
| Headers | token | string | 用户token |
|  Path   |  id   |  int   | TODO id |
返回：

|      字段      |   类型   |    描述     |
|:------------:|:------:|:---------:|
|     code     |  int   |  成功则为200  |
|   message    | string |           |
