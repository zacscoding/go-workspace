# How to use Intellij .http :)  

> My Mock Server  

Start two servers with different port to test environment.

```go
func main() {
	var wait sync.WaitGroup
	wait.Add(2)
	// start server1 (local env)
	go func() {
		e := gin.Default()
		e.GET("/article", func(ctx *gin.Context) {
			log.Println("## Local server")
			getArticleHandle(ctx)
		})
		e.POST("/article", func(ctx *gin.Context) {
			log.Println("## Local server")
			postArticleHandle(ctx)
		})
		if err := e.Run(":3000"); err != nil {
			log.Fatal(err)
		}
		wait.Done()
	}()
	// start server2 (dev env)
	go func() {
		e := gin.Default()
		e.GET("/article", func(ctx *gin.Context) {
			log.Println("## Dev server")
			getArticleHandle(ctx)
		})
		e.POST("/article", func(ctx *gin.Context) {
			log.Println("## Dev server")
			postArticleHandle(ctx)
		})
		if err := e.Run(":3100"); err != nil {
			log.Fatal(err)
		}
		wait.Done()
	}()
	wait.Wait()
}
```  

> basic  

```.http request
### get request with header
GET http://localhost:3000/article
x-request-id: request-id-uuid!!

### post request with payload
POST http://localhost:3000/article
Content-Type: application/json

{
  "title": "Title!!",
  "content": "Content!!",
  "author": {
    "name": "zaccoding!",
    "bio": "It is not working!"
  }
}

### post request with payload read from file
POST http://localhost:3000/article
Content-Type: application/json

< ./article_post.json
```  

> env  

Execute request with env (shortcut : Option + Enter)


```
###
GET {{endpoint}}/article
x-request-id: request-id-uuid!!

###
POST {{endpoint}}/article
Content-Type: application/json

{
  "title": "Title!!",
  "content": "Content!!",
  "author": {
    "name": "zaccoding!",
    "bio": "It is not working!"
  }
}

###
POST {{endpoint}}/article
Content-Type: application/json

< ./article_post.json
```  

> assert response

;TBD