# gonginx

just a tiny toy tool

## Build

1. `go build ./cmd/gonginx/` or `GOOS=linux GOARCH=amd64 go build ./cmd/gonginx/`


## Features

1. listen :8001
2. location ...
3. proxy_pass
4. return
5. echo
6. index root alias
7. default_type

## Configuration

```nginx
# ~/github/gonginx/testdata/a.conf 
http {
    server {
        listen  15001 default_server;
        location ^~ /ESeal/api/ {
           proxy_pass http://127.0.0.1:15002/;
        }
        location / {
            index html/index.html;
            root  .;
        }
	    location ^~ /ESeal/huatai {
           alias /huatai;
           echo "/ESeal/huatai";
	    }
        location /hello { echo "hello, I'am bingoohuang"; }
     }

     server {
        listen  15002 default_server;
        location / { index html/index.html; }
        location /demo { echo $request; }
     }
}
```

## run

```bash
$./gonginx -c ~/github/gonginx/testdata/a.conf                                                                                                                                                                 [五  6/12 11:26:51 2020]
INFO[0000] listening on 15002
INFO[0000] listening on 15001
INFO[0003] received request GET /hello HTTP/1.1
Host: 127.0.0.1:15001
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
Accept-Encoding: gzip, deflate, br
Accept-Language: zh-CN,zh;q=0.9
Cache-Control: max-age=0
Connection: keep-alive
Dnt: 1
Sec-Fetch-Dest: document
Sec-Fetch-Mode: navigate
Sec-Fetch-Site: none
Sec-Fetch-User: ?1
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36

INFO[0021] received request GET /html/index.html HTTP/1.1
Host: 127.0.0.1:15001
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
Accept-Encoding: gzip, deflate, br
Accept-Language: zh-CN,zh;q=0.9
Cache-Control: max-age=0
Connection: keep-alive
Dnt: 1
If-Modified-Since: Mon, 05 Aug 2019 06:32:34 GMT
Sec-Fetch-Dest: document
Sec-Fetch-Mode: navigate
Sec-Fetch-Site: none
Sec-Fetch-User: ?1
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36
```
