package rest

import (
	"github.com/xujiyou-drift/drift/pkg/rest/first"
	"github.com/xujiyou-drift/drift/pkg/rest/kafka"
	"github.com/xujiyou-drift/drift/pkg/rest/zookeeper"
	"log"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "id"

type User struct {
	UserName  string
	FirstName string
	LastName  string
}

func StartRestServer(m manager.Manager) {
	first.Mgr = m
	zookeeper.Mgr = m
	kafka.Mgr = m
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				UserName: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := loginVals.Username
			password := loginVals.Password

			if (userID == "admin" && password == "admin") || (userID == "test" && password == "test") {
				return &User{
					UserName:  userID,
					LastName:  "Bo-Yi",
					FirstName: "Wu",
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*User); ok && v.UserName == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	router.POST("/api/login", authMiddleware.LoginHandler)

	router.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	authApi := router.Group("/api")
	authApi.GET("/api/refresh_token", authMiddleware.RefreshHandler)
	authApi.Use(authMiddleware.MiddlewareFunc())
	{
		authApi.GET("/init", first.FindDriftInitCr)
		authApi.POST("/init", first.CreateDriftInit)
		authApi.POST("/init/pvc", first.RecordPvc)
		authApi.POST("/init/zookeeper", first.CreateZooKeeper)
		authApi.POST("/init/kafka", first.CreateKafka)
		authApi.POST("/init/config", first.CompleteConfig)
		authApi.POST("/init/complete", first.Complete)

		authApi.POST("/zookeeper/status", zookeeper.FindStatus)
		authApi.POST("/kafka/status", kafka.FindStatus)
	}

	if err := http.ListenAndServe("0.0.0.0:8000", router); err != nil {
		log.Fatal(err)
	}
}
