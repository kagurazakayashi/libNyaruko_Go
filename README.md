# libNyaruko_Go
雅诗自用编程封装代码（ Golang 版）

## nyaredis
Redis 数据库相关常用操作的封装。

### 使用
1. 执行下载: `go get github.com/kagurazakayashi/libNyaruko_Go/nyaredis`
2. 代码引入: `import nyaredis "github.com/kagurazakayashi/libNyaruko_Go/nyaredis"`
3. 创建类实例: `nredis = nyaredis.New(confs)`
4. 可以使用 `nredis.任何代码中大写字母开头的方法` ，具体使用方式见源码中每个方法的注释。