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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func UsersRegister(router *gin.RouterGroup) {

	router.POST("/post_add", UsersAddPost)
	router.POST("/list/:id/edit", UpdateUsers)
	router.PUT("/api/list/:id/edit", UpdateUsers)
	router.GET("/list/:id/delete", DeleteUsers)
	router.GET("/add", ViewAddUsers)
	router.GET("/list", ViewUsersList)
	router.GET("/list/:id", ViewUsersListPrev)
	router.GET("/api/list", ApiViewUsersList)
	router.GET("/api/add", ApiViewAddUsers)
	router.GET("/api/list/:id", ApiViewUsersListPrev)
	router.DELETE("/api/list/:id/delete", ApiDeleteUsers)

}

type UsersValidation struct {
	Name     string `form:"name" json:"name" binding:"required,alphanum,min=4,max=255"`
	Email    string `form:"email" json:"email" binding:"required,email"`
	Role     string `form:"role" json:"role"`
	Password string `form:"password" json:"password"`
}

func UsersAddPost(c *gin.Context) {
	// Validate input
	var input UsersValidation

	if err := c.ShouldBind(&input); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	bytePassword := []byte(input.Password)
	// Make sure the second param `bcrypt generator cost` between [4, 32)
	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

	input.Password = string(passwordHash)

	// Create user
	user := Model.UserModel{Name: input.Name, Role: input.Role, Email: input.Email, Password: input.Password}

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

	//c.JSON(http.StatusOK, gin.H{"data": user})

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {

		c.Redirect(http.StatusFound, "/users/list")

	} else {

		c.IndentedJSON(http.StatusCreated, user)

	}

}

func UpdateUsers(c *gin.Context) {
	// Get model if exist

	// Validate input
	var input UsersValidation

	if err := c.ShouldBind(&input); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	if len(input.Password) > 0 {

		bytePassword := []byte(input.Password)
		// Make sure the second param `bcrypt generator cost` between [4, 32)
		passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

		input.Password = string(passwordHash)

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

		objectid, input_id := primitive.ObjectIDFromHex(c.Param("id"))

		filter := bson.M{"_id": objectid}

		update := bson.D{

			{"$set", bson.D{
				{"name", input.Name},
				{"email", input.Email},
				{"role", input.Role},
				{"password", input.Password},
			}},
		}

		_, input_id = collection.UpdateOne(ctx, filter, update)

		if input_id != nil {

			c.JSON(http.StatusBadRequest, gin.H{

				"msg": "err collections find one",
			})

		}
		//end MongoDB

	} else {

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

		objectid, input_id := primitive.ObjectIDFromHex(c.Param("id"))

		filter := bson.M{"_id": objectid}

		update := bson.D{

			{"$set", bson.D{
				{"name", input.Name},
				{"email", input.Email},
				{"role", input.Role},
			}},
		}

		_, input_id = collection.UpdateOne(ctx, filter, update)

		if input_id != nil {

			c.JSON(http.StatusBadRequest, gin.H{

				"msg": "err collections find one",
			})
		}
		//end MongoDB

	}

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {

		c.Redirect(http.StatusFound, "/users/list")

	} else {

		c.IndentedJSON(http.StatusOK, "ok")

	}

}

func DeleteUsers(c *gin.Context) {
	// Get model if exist

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

	objectid, input := primitive.ObjectIDFromHex(c.Param("id"))

	filter := bson.M{"_id": objectid}

	_, input = collection.DeleteMany(ctx, filter)

	if input != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}
	//end MongoDB

	//c.JSON(http.StatusOK, gin.H{"data": true})
	c.Redirect(http.StatusFound, "/users/list")
}

func ViewUsersList(c *gin.Context) {

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {

		//c.JSON(http.StatusForbidden, gin.H{
		//	"message": "not authed",
		//})
		c.Redirect(http.StatusFound, "/auth/login")

		c.Abort()

	}

	//env
	env := godotenv.Load()

	if env != nil {

		panic("Error loading .env file")

	}

	template := os.Getenv("TEMPLATE")

	switch {

	case template == "vue":

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	case template == "html":

		//MongoDB
		filter := bson.M{}

		//// Here's an array in which you can store the decoded documents
		var model []*Model.UserModel

		// Passing nil as the filter matches all documents in the collection

		DB_DATABASE := os.Getenv("DB_DATABASE")

		collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		cur, err := collection.Find(ctx, filter)

		if err != nil {

			log.Fatal(err)

		}

		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		for cur.Next(ctx) {

			// create a value into which the single document can be decoded
			var elem Model.UserModel

			err := cur.Decode(&elem)

			if err != nil {

				log.Fatal(err)
			}

			model = append(model, &elem)

		}

		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}

		// Close the cursor once finished
		cur.Close(ctx)
		//end MongoDB

		//HTML template
		c.HTML(http.StatusOK, "admin_views_users_list.html", gin.H{"csrf": csrf.GetToken(c), "session_id": sessionID, "session_name": sessionName, "list": model})

	default:

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	}

}

func ViewUsersListPrev(c *gin.Context) { // Get model if exist

	var model Model.UserModel

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		//c.JSON(http.StatusForbidden, gin.H{
		//	"message": "not authed",
		//})

		c.Redirect(http.StatusFound, "/auth/login")

		c.Abort()
	}
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

	objectid, input := primitive.ObjectIDFromHex(c.Param("id"))

	filter := bson.M{"_id": objectid}

	input = collection.FindOne(ctx, filter).Decode(&model)
	//errmongo := collection.Find(filter)

	if input != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}

	//end MongoDB

	template := os.Getenv("TEMPLATE")

	switch {

	case template == "vue":

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	case template == "html":

		//HTML template
		c.HTML(http.StatusOK, "admin_views_users_list_prev.html", gin.H{"csrf": csrf.GetToken(c), "session_id": sessionID, "session_name": sessionName, "id": model.ID, "name": model.Name,
			"email": model.Email, "role": model.Role})

	default:

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	}

}

func ViewAddUsers(c *gin.Context) { // Get model if exist

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		//c.JSON(http.StatusForbidden, gin.H{
		//	"message": "not authed",
		//})
		c.Redirect(http.StatusFound, "/auth/login")

		c.Abort()

	}

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
		c.HTML(http.StatusOK, "admin_views_users_add.html", gin.H{"csrf": csrf.GetToken(c), "session_id": sessionID, "session_name": sessionName})

	default:

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	}

}

func ApiViewUsersList(c *gin.Context) {

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		//c.JSON(http.StatusForbidden, gin.H{
		//	"message": "not authed",
		//})

		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_auth_login"})

		c.Abort()

	}

	//MongoDB
	filter := bson.M{}

	//// Here's an array in which you can store the decoded documents
	var model []*Model.UserModel

	// Passing nil as the filter matches all documents in the collection
	//env
	env := godotenv.Load()

	if env != nil {

		panic("Error loading .env file")

	}

	DB_DATABASE := os.Getenv("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	cur, err := collection.Find(ctx, filter)

	if err != nil {

		log.Fatal(err)

	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(ctx) {

		// create a value into which the single document can be decoded
		var elem Model.UserModel

		err := cur.Decode(&elem)

		if err != nil {

			log.Fatal(err)
		}

		model = append(model, &elem)

	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	cur.Close(ctx)
	//end MongoDB

	c.IndentedJSON(http.StatusOK, gin.H{"csrf": csrf.GetToken(c), "session_id": sessionID, "session_name": sessionName, "list": model})

}

func ApiViewAddUsers(c *gin.Context) { // Get model if exist

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		//c.JSON(http.StatusForbidden, gin.H{
		//	"message": "not authed",
		//})

		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_auth_login"})

		c.Abort()

	}

	//c.JSON(http.StatusOK, gin.H{"data": model})
	c.IndentedJSON(http.StatusOK, gin.H{"csrf": csrf.GetToken(c), "session_id": sessionID, "session_name": sessionName})

}

func ApiViewUsersListPrev(c *gin.Context) { // Get model if exist

	var model Model.UserModel

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		//c.JSON(http.StatusForbidden, gin.H{
		//	"message": "not authed",
		//})

		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_auth_login"})

		c.Abort()
	}
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

	objectid, input := primitive.ObjectIDFromHex(c.Param("id"))

	filter := bson.M{"_id": objectid}

	input = collection.FindOne(ctx, filter).Decode(&model)
	//errmongo := collection.Find(filter)

	if input != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}

	//end MongoDB

	//c.JSON(http.StatusOK, gin.H{"data": model })
	c.IndentedJSON(http.StatusOK, gin.H{"csrf": csrf.GetToken(c), "session_id": sessionID, "session_name": sessionName, "id": model.ID, "name": model.Name,
		"email": model.Email, "role": model.Role})

}

func ApiDeleteUsers(c *gin.Context) {
	// Get model if exist

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

	objectid, input := primitive.ObjectIDFromHex(c.Param("id"))

	filter := bson.M{"_id": objectid}

	_, input = collection.DeleteMany(ctx, filter)

	if input != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}
	//end MongoDB

	c.IndentedJSON(http.StatusOK, gin.H{"data": true})
}
