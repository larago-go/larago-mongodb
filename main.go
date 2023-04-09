package main

import (
	"larago/app/Http/Controllers"
	"larago/app/Http/Middleware"
	"larago/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"

	///sessions_redis
	//"github.com/gin-contrib/sessions/redis"
	//end_sessions_redis
	//sessions_cookie
	"github.com/gin-contrib/sessions/cookie"
	//end_sessions_cookie
	//Memcached
	//"github.com/bradfitz/gomemcache/memcache"
	//"github.com/gin-contrib/sessions/memcached"
	//end_Memcached
)

func main() {

	//Mongodb
	config.Init_Mongo()
	//end_Mongodb

	APP_KEYS := config.EnvFunc("APP_KEYS")

	//gin
	//switch to "release" mode in production

	//gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	//Trusted_proxies
	//r.SetTrustedProxies([]string{"192.168.1.2"})
	//end_Trusted_proxies

	//sessions

	//sessions_cookie
	store := cookie.NewStore([]byte(APP_KEYS))
	//end_sessions_cookie

	//redis_sessions
	//REDIS_HOST := config.EnvFunc("REDIS_HOST")
	//REDIS_PASSWORD := config.EnvFunc("REDIS_PASSWORD")
	//REDIS_PORT := config.EnvFunc("REDIS_PORT")
	//REDIS_SECRET := config.EnvFunc("REDIS_SECRET")

	//store, err := redis.NewStore(10, "tcp", REDIS_HOST+":"+REDIS_PORT, REDIS_PASSWORD, []byte(REDIS_SECRET))

	//if err != nil {
	//    panic("Failed to connect to redis_sessions!")
	//    }
	//redis_sessions

	//Memcached
	//store := memcached.NewStore(memcache.New("localhost:11211"), "", []byte("APP_KEYS"))
	//end_Memcached

	//sessions_use
	r.Use(sessions.Sessions(config.EnvFunc("SESSION_NAME"), store))
	//end_sessions

	//CSRF_Middleware
	r.Use(csrf.Middleware(csrf.Options{
		Secret: APP_KEYS,
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))
	//end_CSRF_Middleware

	//gin_html_and_static
	r.Static("/assets", "./dist/assets")
	r.Static("/public", "./public")
	r.Static("/node_modules", "./node_modules")

	//switch vue | html template"
	//html
	//r.LoadHTMLGlob("resources/views/*.html")
	//vue
	r.LoadHTMLGlob("dist/*.html")
	//end_gin_html_and_static

	//gin_route_middleware

	welcome := r.Group("/")
	Controllers.Welcome(welcome.Group("/"))

	auth := r.Group("/auth")
	Controllers.Auth(auth.Group("/"))

	res_pass := r.Group("/login")
	Controllers.Res_pass(res_pass.Group("/"))

	//Auth_Middleware
	r.Use(Middleware.AuthMiddleware(true))
	//end_Auth_Middleware

	home := r.Group("/home")
	Controllers.Home(home.Group("/"))

	//Casbin_Role_Middleware
	//r.Use(Middleware.AuthCasbinMiddleware(true))
	//end_Casbin_Role_Middleware

	users := r.Group("/users")
	Controllers.UsersRegister(users.Group("/"))

	role := r.Group("/role")
	Controllers.CasbinRole(role.Group("/"))

	//end_gin_route_middleware

	//test
	//test := r.Group("/api/ping")
	//test.Use(Middleware.AuthMiddleware(true))

	//test.GET("/", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "pong",
	//	})
	//})
	//end_test

	PORT := config.EnvFunc("PORT")
	r.Run(PORT) // listen and serve on 0.0.0.0:8080
}
