package Controllers

import (
	"context"
	"larago/app/Model"
	"larago/config"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

	input.Password = string(passwordHash)

	// Create user
	user := Model.UserModel{
		Name:     input.Name,
		Role:     input.Role,
		Email:    input.Email,
		Password: input.Password,
	}

	//MongoDB

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

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

	index := mongo.IndexModel{
		Keys:    bson.M{"name": 1},
		Options: opt}

	if _, err := collection.Indexes().CreateOne(ctx, index); err != nil {
		log.Println("Could not create index:", err)
	}

	//end MongoDB

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {
		c.Redirect(http.StatusFound, "/users/list")
	} else {
		c.IndentedJSON(http.StatusCreated, user)
	}

}

func UpdateUsers(c *gin.Context) {

	// Validate input
	var input UsersValidation

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.Password) > 0 {

		bytePassword := []byte(input.Password)

		passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

		input.Password = string(passwordHash)
		//MongoDB

		DB_DATABASE := config.EnvFunc("DB_DATABASE")

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

		_, input_id = collection.UpdateOne(
			ctx,
			filter,
			update)

		if input_id != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "err collections find one",
			})
		}

	} else {

		//MongoDB
		DB_DATABASE := config.EnvFunc("DB_DATABASE")

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

		_, input_id = collection.UpdateOne(
			ctx,
			filter,
			update)

		if input_id != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "err collections find one",
			})
		}

	}

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {
		c.Redirect(http.StatusFound, "/users/list")
	} else {
		c.IndentedJSON(http.StatusOK, "ok")
	}

}

func DeleteUsers(c *gin.Context) {

	//MongoDB
	DB_DATABASE := config.EnvFunc("DB_DATABASE")

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

	c.Redirect(http.StatusFound, "/users/list")
}

func ViewUsersList(c *gin.Context) {

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		c.Redirect(http.StatusFound, "/auth/login")
		c.Abort()
	}

	//env

	template := config.EnvFunc("TEMPLATE")

	switch {
	case template == "vue":
		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})
	case template == "html":
		//MongoDB
		filter := bson.M{}

		var model []*Model.UserModel

		DB_DATABASE := config.EnvFunc("DB_DATABASE")

		collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		cur, err := collection.Find(ctx, filter)

		if err != nil {
			log.Fatal(err)
		}

		for cur.Next(ctx) {

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

		cur.Close(ctx)

		//HTML template
		c.HTML(http.StatusOK, "admin_views_users_list.html", gin.H{
			"csrf":         csrf.GetToken(c),
			"session_id":   sessionID,
			"session_name": sessionName,
			"list":         model})

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
		c.Redirect(http.StatusFound, "/auth/login")
		c.Abort()
	}
	//MongoDB
	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	objectid, input := primitive.ObjectIDFromHex(c.Param("id"))

	filter := bson.M{"_id": objectid}

	input = collection.FindOne(ctx, filter).Decode(&model)

	if input != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "err collections find one",
		})
	}

	template := config.EnvFunc("TEMPLATE")

	switch {
	case template == "vue":
		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})
	case template == "html":
		//HTML template
		c.HTML(http.StatusOK, "admin_views_users_list_prev.html", gin.H{
			"csrf":         csrf.GetToken(c),
			"session_id":   sessionID,
			"session_name": sessionName,
			"id":           model.ID,
			"name":         model.Name,
			"email":        model.Email,
			"role":         model.Role,
		})
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
		c.Redirect(http.StatusFound, "/auth/login")
		c.Abort()
	}

	//env

	template := config.EnvFunc("TEMPLATE")

	switch {
	case template == "vue":
		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})
	case template == "html":
		//HTML template
		c.HTML(http.StatusOK, "admin_views_users_add.html", gin.H{
			"csrf":         csrf.GetToken(c),
			"session_id":   sessionID,
			"session_name": sessionName})
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
		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_auth_login"})
		c.Abort()
	}

	//MongoDB
	filter := bson.M{}

	var model []*Model.UserModel

	//env
	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	cur, err := collection.Find(ctx, filter)

	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(ctx) {

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

	cur.Close(ctx)

	c.IndentedJSON(http.StatusOK, gin.H{
		"csrf":         csrf.GetToken(c),
		"session_id":   sessionID,
		"session_name": sessionName,
		"list":         model})

}

func ApiViewAddUsers(c *gin.Context) { // Get model if exist

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_auth_login"})
		c.Abort()
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"csrf":         csrf.GetToken(c),
		"session_id":   sessionID,
		"session_name": sessionName})

}

func ApiViewUsersListPrev(c *gin.Context) { // Get model if exist

	var model Model.UserModel

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_auth_login"})
		c.Abort()
	}
	//MongoDB
	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	objectid, input := primitive.ObjectIDFromHex(c.Param("id"))

	filter := bson.M{"_id": objectid}

	input = collection.FindOne(ctx, filter).Decode(&model)

	if input != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "err collections find one",
		})
	}

	//end MongoDB

	c.IndentedJSON(http.StatusOK, gin.H{
		"csrf":         csrf.GetToken(c),
		"session_id":   sessionID,
		"session_name": sessionName,
		"id":           model.ID,
		"name":         model.Name,
		"email":        model.Email,
		"role":         model.Role,
	})

}

func ApiDeleteUsers(c *gin.Context) {

	//MongoDB
	DB_DATABASE := config.EnvFunc("DB_DATABASE")

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

	c.IndentedJSON(http.StatusOK, gin.H{"data": true})
}
