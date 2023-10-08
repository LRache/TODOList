# TODOList程序介绍

## 数据库结构  
### SQL  
#### 表USERS

|   字段   |    类型     |        描述        |
| :------: | :---------: | :----------------: |
|    id    |     int     |     用户唯一ID     |
| username | varchar(64) |      用户昵称      |
| password | varchar(64) | 用户密码，加密存储 |
| mailAddr | varchar(64) |      用户邮箱      |

#### 表TODO

|    字段     |   类型   |          描述          |
| :---------: | :------: | :--------------------: |
|     id      |   int    | 每个用户中TODO唯一的ID |
|    title    |   text   |        TODO标题        |
|   content   |   text   |        TODO内容        |
| create_time | datetime |      TODO创建时间      |
|  deadline   | datetime |      TODO截止时间      |
|     tag     |   text   |        TODO标签        |
|   userid    |   int    |    TODO所属用户的ID    |
|    done     |   bool   |     TODO是否已完成     |
|    keyid    |   int    |    自增，TODO唯一ID    |

### REDIS

|         字段         |  类型  |                             描述                             |
| :------------------: | :----: | :----------------------------------------------------------: |
|      UserCount       | string |                           用户总数                           |
|     EmptyUserId      |  list  |              空置用户ID列表，注册用户时优先取用              |
|      ItemCount       |  hash  |                每个用户的TODO数量，键是用户ID                |
| EmptyItemId:`userid` |  list  |       ID对应用户的空置TODO ID列表，添加TODO时优先取用        |
|    MailVerifyCode    |  hash  | 邮箱对应验证码，存储的验证码由验证码和过期时间组成，服务器定期清理过期验证码，键是邮箱地址 |
|    UserTokenCode     |  hash  |             用户id对应随机生成的用户Token验证码              |

## 程序结构、实现功能

### 包

* **globals** 存放各种常量，如返回的Json值、内部错误代码等，以及全局变量，如数据库、随机数引擎、邮件拨号对象等。
* **handler** 作为中间件处理token。
* **model** 定义各种模型结构体，如接受添加Todo请求时使用的`RequestTodoItemModel`，接受修改Todo请求时使用的`RequestUpdateTodoItemModel`，数据库中存放Todo时使用的`ataBaseTodoItemModel`，处理用户token的`UserClaimsModel`模型，以及提供了各种模型之间的转换。
* **server** 核心包，负责接受请求、处理数据、直接操作数据库。
* **utils** 实用工具函数包，如判断用户名是否合法、快捷生成Token等。

### 用户处理部分

#### Token

用户的Token分为用于验证身份的Token和用于刷新用于验证身份Token的Token，他们的过期时间不同。同时，每个用户都有唯一随机的Token验证码，在用户进行修改密码等操作后，该验证码会改变，使原来的Token无效。Token验证码存储在Redis数据库中。

#### 用户邮箱认证

用户以邮箱地址标识。用户可以请求服务器发送验证码到邮箱，再将验证码与邮箱地址发送给服务器验证，验证成功后服务器会发送给用户用于邮箱验证的Token，用户可以凭借这个Token进行注册、修改密码等操作。

邮箱验证码被存储在Redis数据库中，其中前六位为验证码，后面为过期时间，服务器每隔一分钟清理过期验证码。

### Todo Item处理部分

程序实现了增删改查操作。

#### 增加、删除、更新

程序直接操作数据库运行数据库命令实现功能。其中更新字段由前段发送的`updateKeys`字段控制。

#### 查找

程序支持用户直接通过每个Todo的唯一ID查找，也可以进行筛选，查询多个Todo，通过数据库的`SELECT`语句实现。可以查询`done=true`来返回已经完成的todo，通过`deadlineBefore`参数可以查询在一定时间前过期的Todo。

返回多个Todo时，支持设定排序分页查询。

#### 定时提醒

程序通过cron包实现针对Todo的定时函数。定时提醒功能还需要前段配合。

### ID分配

每个用户、Todo（对于每个用户）都有唯一的ID，在用户或者Todo被删除后，留空的ID会被存储在Redis数据库中，下次添加用户或者Todo时会优先从列表中获取留空ID。

