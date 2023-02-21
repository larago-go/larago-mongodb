package Controllers

import (
	"context"
	"larago/app/Model"
	"larago/config"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func Res_pass(router *gin.RouterGroup) {
	router.POST("/post_add", PostForgotPassword)
	router.GET("/forgot_password", ViewForgotPassword)
	router.POST("/pass/:url/post", ViewRes_passListPost)
	router.GET("/pass/:url", ViewRes_passListPrev)
	router.GET("/api/pass/:url", ApiViewRes_passListPrev)
	router.GET("/api/forgot_password", ApiViewForgotPassword)
}

type Res_passValidation struct {
	Email string `form:"email" json:"email" binding:"required,email"`
}

type Res_passPasswordValidation struct {
	Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
}

func PostForgotPassword(c *gin.Context) {
	// Validate input
	var input Res_passValidation

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var model Model.UserModel

	//MongoDB
	//env

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection_users := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx_users, cancel_users := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel_users()

	filter_users := bson.M{"email": input.Email}

	decode := collection_users.FindOne(ctx_users, filter_users).Decode(&model)

	if decode != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}

	rand_urls := config.RandomString(90)

	//smtp - forgot_password

	toList := []string{input.Email}

	body := []byte("From:" + config.EnvFunc("MAIL_USERNAME") + "\r\n" +
		"To:" + input.Email + "\r\n" +
		"Subject: Password recovery\r\n\r\n" +
		"Link to create a new password" + " " + config.EnvFunc("WWWROOT") + "/login/pass/" + rand_urls + "\r\n")

	auth := smtp.PlainAuth("", config.EnvFunc("MAIL_USERNAME"), config.EnvFunc("MAIL_PASSWORD"), config.EnvFunc("MAIL_HOST"))

	smtp.SendMail(config.EnvFunc("MAIL_HOST")+":"+config.EnvFunc("MAIL_PORT"), auth, config.EnvFunc("MAIL_USERNAME"), toList, body)

	//err := smtp.SendMail(config.EnvFunc("MAIL_HOST")+":"+config.EnvFunc("MAIL_PORT"), auth, config.EnvFunc("MAIL_USERNAME"), toList, body)

	// handling the errors
	//if err != nil {
	//  fmt.Println(err)
	//  os.Exit(1)
	//}

	//Gorm_SQL

	url_res := Model.ResPassUserModel{Email: input.Email, Url_full: config.EnvFunc("WWWROOT") + "/login/pass/" + rand_urls, Url: rand_urls}

	collection_respass := config.MongoClient.Database(DB_DATABASE).Collection("respassusermodels")

	ctx_url, cancel_url := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel_url()

	_, err_post := collection_respass.InsertOne(ctx_url, url_res)

	if err_post != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "A user with the same name already exists",
		})

	}

	opt := options.Index()

	opt.SetUnique(true)

	index := mongo.IndexModel{Keys: bson.M{"url": 1}, Options: opt}

	if _, err_index := collection_respass.Indexes().CreateOne(ctx_url, index); err_index != nil {

		log.Println("Could not create index:", err_index)

	}

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {

		c.Redirect(http.StatusFound, "/")

	} else {

		c.IndentedJSON(http.StatusOK, gin.H{"data": true})

	}

	//remove link password recovery after 30 minutes

	time.AfterFunc(30*time.Minute, func() {

		collection_respass_del := config.MongoClient.Database(DB_DATABASE).Collection("respassusermodels")

		ctx_respass, cancel_respass := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel_respass()

		filter_respass_del := bson.M{"email": input.Email}

		_, err_respass_del := collection_respass_del.DeleteMany(ctx_respass, filter_respass_del)

		if err_respass_del != nil {

			c.JSON(http.StatusBadRequest, gin.H{

				"msg": "err collections find one",
			})

		}

	})
}

func ViewRes_passListPrev(c *gin.Context) { // Get model if exist

	var model Model.ResPassUserModel

	//MongoDB
	//env

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	filter := bson.M{"url": c.Param("url")}

	res := collection.FindOne(ctx, filter).Decode(&model)
	//errmongo := collection.Find(filter)

	if res != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}

	//end MongoDB

	template := config.EnvFunc("TEMPLATE")

	switch {

	case template == "vue":

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	case template == "html":

		//HTML template
		c.HTML(http.StatusOK, "admin_auth_forgot_password_new.html", gin.H{"csrf": csrf.GetToken(c), "url": model.Url})

	default:

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	}

}

func ViewRes_passListPost(c *gin.Context) { // Get model if exist

	var model Model.ResPassUserModel

	//MongoDB
	//env

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection_respass := config.MongoClient.Database(DB_DATABASE).Collection("respassusermodels")

	ctx_respass, cancel_respass := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel_respass()

	filter_respass := bson.M{"url": c.Param("url")}

	decode_respass := collection_respass.FindOne(ctx_respass, filter_respass).Decode(&model)

	if decode_respass != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}

	//end MongoDB

	var input Res_passPasswordValidation

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bytePassword := []byte(input.Password)
	// Make sure the second param `bcrypt generator cost` between [4, 32)
	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

	input.Password = string(passwordHash)

	collection_users := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx_users, cancel_users := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel_users()

	filter_users := bson.M{"email": model.Email}

	update_users := bson.D{

		{"$set", bson.D{
			{"password", input.Password},
		}},
	}

	_, err_users := collection_users.UpdateOne(ctx_users, filter_users, update_users)

	if err_users != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}
	//end MongoDB

	//c.JSON(http.StatusOK, gin.H{"data": model })

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {

		c.Redirect(http.StatusFound, "/auth/login")

	} else {

		c.IndentedJSON(http.StatusOK, gin.H{"data": true})

	}
}

func ViewForgotPassword(c *gin.Context) { // Get model if exist

	//env

	template := config.EnvFunc("TEMPLATE")

	switch {

	case template == "vue":

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	case template == "html":

		//HTML template
		c.HTML(http.StatusOK, "admin_auth_forgot_password.html", gin.H{"csrf": csrf.GetToken(c)})

	default:

		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	}

}

func ApiViewForgotPassword(c *gin.Context) { // Get model if exist

	c.IndentedJSON(http.StatusOK, gin.H{"csrf": csrf.GetToken(c)})
}

func ApiViewRes_passListPrev(c *gin.Context) { // Get model if exist

	var model Model.ResPassUserModel

	//MongoDB
	//env

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("usermodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	filter := bson.M{"url": c.Param("url")}

	res := collection.FindOne(ctx, filter).Decode(&model)
	//errmongo := collection.Find(filter)

	if res != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"msg": "err collections find one",
		})

	}

	//end MongoDB
	//c.JSON(http.StatusOK, gin.H{"data": model })
	c.IndentedJSON(http.StatusOK, gin.H{"csrf": csrf.GetToken(c), "url": model.Url})
}
