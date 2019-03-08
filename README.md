# consul-go-grpc-demo
用consul做grpc的服务注册与发现，示例代码

### requires:
go 1.12

### run
* run consul

```bash
consul agent -dev
```
   

* server

```bash
cd server
go mod tidy
go run cmd/main.go
```

    
  
* client 

```bash
cd client
go mod tidy
go run cmd/main.go
```

    
