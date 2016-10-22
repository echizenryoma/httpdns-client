# qdns

一个使用GO语言编写的使用腾讯HTTP DNS [<http://pdns.dnspod.cn>] 的客户端。

## 编译

```
go build -ldflags "-s -w" .
```

## 用法

```
-dns string
      Upstream DNS Server Address (default "119.29.29.29")
-httpdns string
      Tencent HTTP DNS address (default "119.29.29.29")
-ip string
      DNS bind IP address (default "127.0.0.1")
-port int
      Listen on port (default 53)
-save
      Whether to save the results to a sqlite (default true)
-workers int
      number of independent workers (default 10)
```


## 依赖

```
github.com/miekg/dns
github.com/mattn/go-sqlite3
github.com/patrickmn/go-cache
golang.org/x/net/publicsuffix
```