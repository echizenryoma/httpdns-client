# qdns

一个使用GO语言编写的使用腾讯HTTP DNS [<http://pdns.dnspod.cn>] 的客户端。

## 编译

```
go build -ldflags "-s -w" .
```

## 用法

```
Usage of qdns.exe:
  -db
        Save result to sqlite (default true)
  -ip string
        DNS bind IP address (default "127.0.0.1")
  -path string
        Sqlite path (default "insight.sqlite")
  -port int
        listen on port number (default 53)
  -server string
        Tencent HTTP DNS address (default "119.29.29.29")
  -workers int
        number of independent workers (default 10)
```
