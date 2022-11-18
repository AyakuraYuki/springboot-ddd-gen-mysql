# SpringBoot DDD gen mysql

读取 MySQL 表信息，生成 DDD 模式的，可以在 SpringBoot 项目中工作的代码。

目前生成的代码仅适用于公司的项目结构，可以按需调整内容。

## install

```shell
go install -v github.com/AyakuraYuki/springboot-ddd-gen-mysql@latest

$GOBIN/springboot-ddd-gen-mysql -h
# -h string
#       host - 主机名，默认 localhost (default "localhost")
# -P int
#       port - 端口号，默认 3306 (default 3306)
# -u string
#       user - 用户名 (default "root")
# -p string
#       password - 密码 (default "root")
# -d string
#       schema name - 数据库名 (default "db_local")
# -t string
#       table name - 表名 (default "tb_user")
# -D string
#       domain name - 领域名 (default "user")
```

## build from source codes

1. clone this repository
2. enter the path to your clone
3. run cmd `go build -o springboot-ddd-gen-mysql`
