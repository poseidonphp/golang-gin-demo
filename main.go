package main

import (
	"demo/controllers"
	"demo/initializers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
)

var db = make(map[string]string)

func init() {
	// Load the .env file
	initializers.LoadEnvs()

	// Initialize our memory store (Redis)
	initializers.ConnectMemoryDb()

	//initializers.ConnectDB()
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	//r.POST("/auth/signup", controllers.CreateUser)
	//r.POST("/auth/login", controllers.Login)
	//r.GET("/user/profile", middlewares.CheckAuth, controllers.GetUserProfile)

	// EXAMPLES USING SESSIONS

	// Initialize the session handler middleware which will use our memory store
	s := initializers.KvStore
	sessionAge, _ := strconv.Atoi(os.Getenv("SESSION_LIFETIME"))
	s.Options(sessions.Options{
		MaxAge: sessionAge, // 30 seconds for testing
	})
	r.Use(sessions.Sessions(os.Getenv("SESSION_KEY"), s))

	r.GET("/login", controllers.LoginHandler)
	r.GET("/callback", controllers.CallbackHandler)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "hello world"})
	})

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		session := sessions.Default(c)
		var count int
		v := session.Get("count")

		if v == nil {
			count = 0
		} else {
			count = v.(int)
			count++
		}
		session.Set("count", count)
		session.Save()
		c.JSON(200, gin.H{"count": count})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
