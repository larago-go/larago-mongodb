package Controllers

import (
	"context"
	"larago/app/Model"
	"larago/config"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	csrf "github.com/utrack/gin-csrf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func Auth(router *gin.RouterGroup) {

	router.POST("/signup", UsersRegistration)
	router.POST("/signin", UsersLogin)
	router.GET("/signout", Loginout)
	router.GET("/login", ViewUsersLogin)
	router.GET("/register", ViewUsersRegistration)
	router.GET("/api/register", ApiViewUsersRegistration)
	router.GET("/api/login", ApiViewUsersLogin)
	router.GET("/api/session", ViewUserSession)
	router.GET("/api/signout", ApiLoginout)
}

type PasswordValidation struct {
	Name     string `form:"name" json:"name" binding:"required,alphanum,min=4,max=255"`
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
}

type LoginValidation struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password"json:"password" binding:"required,min=8,max=255"`
}

func UsersRegistration(c *gin.Context) {

	// Validate input
	var input PasswordValidation

	if err := c.ShouldBind(&input); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	bytePassword := []byte(input.Password)

	// Make sure the second param `bcrypt generator cost` between [4, 32)
	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

	input.Password = string(passwordHash)

	// Create user
	user := Model.UserModel{Name: input.Name, Email: input.Email, Password: input.Password}

	//MongoDB

	//env
	env := godotenv.Load()

	if env != nil {

		panic("Error loading .env file")

	}

	DB_DATABASE := os.Getenv("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	_, err_post := collection.InsertOne(ctx, user)

	if err_post != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "A user with the same name already exists",
		})

	}

	opt := options.Index()

	opt.SetUnique(true)

	index := mongo.IndexModel{Keys: bson.M{"name": 1}, Options: opt}

	if _, err := collection.Indexes().CreateOne(ctx, index); err != nil {

		log.Println("Could not create index:", err)
	}

	//end MongoDB

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {

		c.Redirect(http.StatusFound, "/home")

	} else {

		c.IndentedJSON(http.StatusCreated, user)

	}

}

func UsersLogin(c *gin.Context) {

	// Validate input
	var input LoginValidation

	if err := c.ShouldBind(&input); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	var model Model.UserModel

	//MongoDB

	//env

	env := godotenv.Load()

	if env != nil {

		panic("Error loading .env file")

	}

	DB_DATABASE := os.Getenv("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	filter := bson.M{"email": input.Email}

	errmongo := collection.FindOne(ctx, filter).Decode(&model)

	if errmongo != nil {

		log.Fatal("err collections users")

	}

	//end MongoDB

	bytePassword := []byte(input.Password)

	byteHashedPassword := []byte(model.Password)

	err := bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Password mismatch",
		})

		return

	} else {

		session := sessions.Default(c)

		session.Set("user_id", model.ID)

		session.Set("user_email", model.Email)

		session.Set("user_name", model.Name)

		//Casbinrole
		session.Set("user_role", model.Role)

		session.Save()

		//c.JSON(http.StatusOK, gin.H{"message": "User signed in", "user": model.Name, "id": model.ID})

		headerContentTtype := c.Request.Header.Get("Content-Type")

		if headerContentTtype != "application/json" {

			c.Redirect(http.StatusFound, "/home")

		} else {

			c.IndentedJSON(http.StatusCreated, gin.H{"message": "User signed in", "user": model.Name, "id": model.ID})

		}

	}

}

func Loginout(c *gin.Context) {

	session := sessions.Default(c)

	session.Clear()

	session.Save()

	c.Redirect(http.StatusFound, "/")

	//c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})

}

func ViewUsersLogin(c *gin.Context) {

	session := sessions.Default(c)

	sessionID := session.Get("user_id")

	if sessionID == nil {

		//c.JSON(http.StatusForbidden, gin.H{
		//  "message": "not authed",
		//})
		//c.Redirect(http.StatusFound, "/auth/login")
		//c.Abort()
		//c.HTML(http.StatusOK, "login.html", gin.H{"csrf": csrf.GetToken(c)})

		//env
		env := godotenv.Load()

		if env != nil {

			panic("Error loading .env file")

		}
		//end_env

		template := os.Getenv("TEMPLATE")

		switch {

		case template == "vue":

			//VUE template
			c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

		case template == "html":

			//HTML template
			c.HTML(http.StatusOK, "admin_auth_login.html", gin.H{"csrf": csrf.GetToken(c)})

		default:

			//VUE template
			c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

		}

	} else {

		c.Redirect(http.StatusFound, "/home")

	}

}

func ViewUsersRegistration(c *gin.Context) {

	session := sessions.Default(c)

	sessionID := session.Get("user_id")

	if sessionID == nil {

		//c.JSON(http.StatusForbidden, gin.H{
		//  "message": "not authed",
		//})
		//c.Redirect(http.StatusFound, "/auth/login")
		//c.Abort()
		//c.HTML(http.StatusOK, "login.html", gin.H{"csrf": csrf.GetToken(c)})

		//env
		env := godotenv.Load()

		if env != nil {

			panic("Error loading .env file")

		}
		//end_env

		template := os.Getenv("TEMPLATE")

		switch {

		case template == "vue":

			//VUE template
			c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

		case template == "html":

			//HTML template
			c.HTML(http.StatusOK, "admin_auth_register.html", gin.H{"csrf": csrf.GetToken(c)})

		default:

			//VUE template
			c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

		}

	} else {

		c.Redirect(http.StatusFound, "/home")

	}

}

func ApiViewUsersRegistration(c *gin.Context) {

	session := sessions.Default(c)

	sessionID := session.Get("user_id")

	if sessionID == nil {

		c.IndentedJSON(http.StatusOK, gin.H{"csrf": csrf.GetToken(c)})

	} else {

		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_home"})

	}

}

func ApiViewUsersLogin(c *gin.Context) {

	session := sessions.Default(c)

	sessionID := session.Get("user_id")

	if sessionID == nil {

		c.IndentedJSON(http.StatusOK, gin.H{"csrf": csrf.GetToken(c)})

	} else {

		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_home"})

	}

}

func ViewUserSession(c *gin.Context) {

	session := sessions.Default(c)

	sessionID := session.Get("user_id")

	if sessionID == nil {

		c.IndentedJSON(http.StatusOK, gin.H{"userid_session_id": "no_auth", "userid_session": "no_auth"})

	} else {

		c.IndentedJSON(http.StatusOK, gin.H{"userid_session_id": sessionID, "userid_session": "auth"})

	}

}

func ApiLoginout(c *gin.Context) {

	session := sessions.Default(c)

	session.Clear()

	session.Save()

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Signed out..."})

}
