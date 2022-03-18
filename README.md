# libNyaruko_Go
- 雅诗编程封装代码 libNyaruko ( Golang 版 )
  - 通用工具库，一些跨项目常用的代码会放在里面。
  - 分为单独的组件，可以按需引入。
  - (前端: [libNyaruko_TS](https://github.com/kagurazakayashi/libNyaruko_TS) )

## 组件
- `nyacrypt`
  - 字符串或文件哈希计算与加密，随机数或随机字符串生成。
- `nyahttphandle`
  - HTTP 服务。用于和客户端 GET/POST 交互，内置多语言信息 JSON 生成器。
- `nyaio`
  - 文件相关功能。获取文件信息，多种方式读取，遍历，新建文件夹等。
- `nyalog`
  - 日志信息输出和记录功能。支持彩色日志信息输出，自动染色，调试信息突出显示，记录到文件等功能。
- `nyamqtt`
  - MQTT 通讯。订阅消息，发布消息，批量订阅和退订等。
- `nyamysql`
  - MySQL 数据库连接，执行 SQL 语句。内置常用操作指令函数。
- `nyaredis`
  - Redis 数据库连接，设置键值，生存时间，通配符查询，批量删除等。
- `nyasql`
  - SQL 语句生成器。快捷模板式生成 SQL 语句，内置安全检查和转义。可用于 `nyamysql` 和 `nyasqlite`
- `nyasqlite`
  - SQLite 数据库连接，执行 SQL 语句。内置常用操作指令函数。

## 使用
- 导入所需要功能所在的文件即可。例如：
  - 下载包: `go get github.com/kagurazakayashi/libNyaruko_Go/nyaredis`
  - 在代码中导入: `import nyaredis "github.com/kagurazakayashi/libNyaruko_Go/nyaredis"`
  - 创建类实例: `nredis = nyaredis.New(confs)`
  - 可以使用 `nredis.任何代码中大写字母开头的函数` ，具体使用方式见 Wiki 或源码中每个函数的注释。

## 详细使用帮助
参考 [Wiki](https://github.com/kagurazakayashi/libNyaruko_TS/wiki) ，在 Wiki 右侧的 Pages 点相关文件名即可了解。

## LICENSE
- [木兰宽松许可证， 第2版](http://license.coscl.org.cn/MulanPSL2)
- [Mulan Permissive Software License，Version 2 (Mulan PSL v2)](http://license.coscl.org.cn/MulanPSL2)
