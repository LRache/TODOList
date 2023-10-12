# TODOList
## 运行
### 初始化数据库
#### MySQL  
```mysql
CREATE DATABASE tododata;
```
```mysql
USE tododata;
CREATE TABLE IF NOT EXISTS Users(
    id INT,  
    username VARCHAR(64), 
    password VARCHAR(64), 
    todocount INT, 
    mailAddr VARCHAR(64)
) CHARSET=utf8;
CREATE TABLE IF NOT EXISTS todo(
    id INT, 
    title TEXT, 
    content TEXT, 
    create_time DATETIME, 
    deadline DATETIME, 
    tag TEXT, 
    userid INT, 
    done BOOL, 
    keyid INT UNSIGNED AUTO_INCREMENT PRIMARY KEY
) CHARSET=utf8;
```
## 数据结构
### TODOItem

|    字段    |  类型  |   描述   |
| :--------: | :----: | :------: |
|     id     |  int   | TODO id  |
|   title    | string |   标题   |
|  content   | string |   内容   |
| createTime |  int   | 创建时间 |
|  deadline  |  int   | 截止时间 |
|    tag     | string |   标签   |
|    done    |  bool  | 是否完成 |

## 接口API
### 发送邮箱验证码
路由：`/todo/user/mail`  
方法：`GET`  
参数：

|  位置   |  字段  |   类型   |  描述  |
|:-----:|:----:|:------:|:----:|
| Query | mail | string | 邮箱地址 |

返回：

|   字段    |   类型   |   描述    |
|:-------:|:------:|:-------:|
|  code   |  int   | 成功则为200 |
| message | string |         |

### 验证邮箱
路由：`/todo/user/mail`  
方法：`POST`  
参数：

|    位置     |  字段  |   类型   |    描述     |
|:---------:|:----:|:------:|:---------:|
| Body Json | mail | string |   用户邮箱    |
| Body Json | code | string | 发送到邮箱的验证码 |

返回：

|    字段     |   类型   |      描述      |
|:---------:|:------:|:------------:|
|   code    |  int   |   成功则为200    |
|  message  | string |              |
| mailToken | string | 用于验证邮箱的token |

### 注册
路由：`/todo/user`  
方法：`PUT`   
参数：

|    位置     |    字段     |   类型   |         描述         |
|:---------:|:---------:|:------:|:------------------:|
| Body Json | mailAddr  | string |        用户邮箱        |
| Body Json | password  | string |        用户密码        |
| Body Json | mailToken | string | 用于确保用户邮箱通过验证的token |

返回：

|      字段      |   类型   |         描述          |
|:------------:|:------:|:-------------------:|
|     code     |  int   |       成功则为201       |
|   message    | string |                     |
|    userId    |  int   |       新用户的id        |
|    token     | string |    用于自动登录的token     |
| refreshToken | string | 用于在token过期后刷新的token |

注意：作为测试，以`@todouser`结尾的邮箱不需要验证。

### 登录和注销
路由： `/todo/user`  
方法：`POST`  
参数：

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

### 设置密码
路由：`/todo/user/reset`  
方法：`POST`
参数：

|    位置     |     字段      |   类型   |         描述         |
|:---------:|:-----------:|:------:|:------------------:|
| Body Json |  mailAddr   | string |        用户邮箱        |
| Body Json | newPassword | string |       用户新的密码       |
| Body Json |  mailToken  | string | 用于确保用户邮箱通过验证的token |

返回：

|      字段      |   类型   |              描述              |
|:------------:|:------:|:----------------------------:|
|     code     |  int   |           成功则为200            |
|   message    | string |                              |


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
| Body Json | createTime |  int   | TODO创建时间 |
| Body Json |  deadline  |  int   | TODO截止时间 |
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

|   位置    |       字段       |   类型   |                  描述                  |
|:-------:|:--------------:|:------:|:------------------------------------:|
| Headers |     token      | string |               用户token                |
|  Query  |      tag       | string |               留空则不筛选标题               |
|  Query  |      done      |  bool  |              留空则不筛选是否完成              |
|  Query  | deadlineBefore |  int   |      筛选截止时间在某时间之前的TODO，留空则表示不筛选      |
|  Query  |   pageIndex    |  int   |               分页查询时的页数               |
|  Query  |     limit      |  int   |             一页中最多TODO数量              |
|  Query  |     order      | string | 排序方式，可选`id` `createTime` `deadline`。 |

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
| Body Json | updateKeys | list[string] | 要更新的字段名称列表 |
| Body Json |   title    |    string    |   TODO标题   |
| Body Json |  content   |    string    |   TODO内容   |
| Body Json | createTime |     int      |  TODO创建时间  |
| Body Json |  deadline  |     int      |  TODO截止时间  |
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
|  Query  |  id   |  int   | TODO id |

返回：

|      字段      |   类型   |    描述     |
|:------------:|:------:|:---------:|
|     code     |  int   |  成功则为200  |
|   message    | string |           |

### 设置TODO定时提醒
路由：`/todo/item/cron`  
方法：`PUT`
参数：

|   位置    |   字段   |   类型   |          描述          |
|:-------:|:------:|:------:|:--------------------:|
| Headers | token  | string |       用户token        |
|  Query  |   id   |  int   |       TODO id        |
|  Query  | before | string | 在多久之前提醒，格式满足Duration |

返回：

|      字段      |   类型   |   描述    |
|:------------:|:------:|:-------:|
|     code     |  int   | 成功则为201 |
|   message    | string |         |

<b>注意：TODO定时提醒函数未实现，后续可以加入向客户端主动发送消息等功能。</b>

## 数据库结构  
### SQL  
#### 表USERS

|    字段    |     类型      |    描述     |
|:--------:|:-----------:|:---------:|
|    id    |     int     |  用户唯一ID   |
| username | varchar(64) |   用户昵称    |
| password | varchar(64) | 用户密码，加密存储 |
| mailAddr | varchar(64) |   用户邮箱    |

#### 表TODO

|     字段      |    类型    |       描述       |
|:-----------:|:--------:|:--------------:|
|     id      |   int    | 每个用户中TODO唯一的ID |
|    title    |   text   |     TODO标题     |
|   content   |   text   |     TODO内容     |
| create_time | datetime |    TODO创建时间    |
|  deadline   | datetime |    TODO截止时间    |
|     tag     |   text   |     TODO标签     |
|   userid    |   int    |  TODO所属用户的ID   |
|    done     |   bool   |   TODO是否已完成    |
|    keyid    |   int    |  自增，TODO唯一ID   |

### REDIS

|          字段          |   类型   |                      描述                       |
|:--------------------:|:------:|:---------------------------------------------:|
|      UserCount       | string |                     用户总数                      |
|     EmptyUserId      |  list  |              空置用户ID列表，注册用户时优先取用               |
|      ItemCount       |  hash  |              每个用户的TODO数量，键是用户ID               |
| EmptyItemId:`userid` |  list  |        ID对应用户的空置TODO ID列表，添加TODO时优先取用         |
|    MailVerifyCode    |  hash  | 邮箱对应验证码，存储的验证码由验证码和过期时间组成，服务器定期清理过期验证码，键是邮箱地址 |
|    UserTokenCode     |  hash  |             用户id对应随机生成的用户Token验证码             |

