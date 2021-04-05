# Learning gRPC with Golang :)  

## Simple  
[source](./simple)  
Simple GRPC Server/Client  

> Generate gRPC code  

```bash
$ protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    simple/hello/hello_service.proto
```  

## Route Guide  
[source](./route)  

> Generate gRPC code  

```bash
$ protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    route/person/person_route.proto
```
