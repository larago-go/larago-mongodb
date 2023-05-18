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
)

func CasbinRole(router *gin.RouterGroup) {

	router.POST("/post_add", AddPostCasbinRole)
	router.GET("/list/:id/delete", DeleteCasbinRole)
	router.GET("/list", ViewCasbinRole)
	router.GET("/add", AddCasbinRole)
	router.GET("/api/add", ApiAddCasbinRole)
	router.GET("/api/list", ApiViewCasbinRole)
	router.DELETE("/api/list/:id/delete", ApiDeleteCasbinRole)

}

type CasbinRoleAddValidation struct {
	RoleName string `form:"rolename" json:"rolename" binding:"required"`
	Path     string `form:"path" json:"path" binding:"required"`
	Method   string `form:"method" json:"method" binding:"required"`
}

func AddPostCasbinRole(c *gin.Context) {
	// Validate input
	var input CasbinRoleAddValidation

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e := config.CasbinRole()

	e.AddPolicy(
		input.RoleName,
		input.Path,
		input.Method,
	)

	// Create role
	role := Model.CasbinRoleModel{
		RoleName: input.RoleName,
		Path:     input.Path,
		Method:   input.Method,
	}

	//MongoDB
	//env

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("casbinrolemodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	_, err_post := collection.InsertOne(ctx, role)

	if err_post != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "A role with the same name already exists",
		})
	}

	opt := options.Index()

	opt.SetUnique(false)

	index := mongo.IndexModel{
		Keys:    bson.M{"v0": 1},
		Options: opt,
	}

	if _, err := collection.Indexes().CreateOne(ctx, index); err != nil {
		log.Println("Could not create index:", err)
	}
	//end MongoDB

	headerContentTtype := c.Request.Header.Get("Content-Type")

	if headerContentTtype != "application/json" {
		c.Redirect(http.StatusFound, "/role/list")
	} else {
		c.IndentedJSON(http.StatusCreated, role)
	}

}

func ViewCasbinRole(c *gin.Context) {

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
		var model []*Model.CasbinRoleModel
		DB_DATABASE := config.EnvFunc("DB_DATABASE")
		collection := config.MongoClient.Database(DB_DATABASE).Collection("casbinrolemodels")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cur, err := collection.Find(ctx, filter)

		if err != nil {
			log.Fatal(err)
		}

		for cur.Next(ctx) {
			var elem Model.CasbinRoleModel
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
		c.HTML(http.StatusOK, "admin_views_casbin_role.html", gin.H{
			"session_id":   sessionID,
			"session_name": sessionName,
			"list":         model,
		})

	default:
		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})

	}

}

func AddCasbinRole(c *gin.Context) {

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
		c.HTML(http.StatusOK, "admin_views_casbin_role_add.html", gin.H{
			"csrf":         csrf.GetToken(c),
			"session_id":   sessionID,
			"session_name": sessionName,
		})

	default:
		//VUE template
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Larago"})
	}

}

func DeleteCasbinRole(c *gin.Context) {

	var model Model.CasbinRoleModel

	//MongoDB
	//env

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("casbinrolemodels")

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

	e := config.CasbinRole()

	e.RemovePolicy(
		model.RoleName,
		model.Path,
		model.Method,
	)

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "err collections find one",
		})
	}

	c.Redirect(http.StatusFound, "/role/list")
}

func ApiViewCasbinRole(c *gin.Context) {

	session := sessions.Default(c)
	sessionID := session.Get("user_id")
	sessionName := session.Get("user_name")

	if sessionID == nil {
		c.IndentedJSON(http.StatusOK, gin.H{"csrf": "redirect_auth_login"})
		c.Abort()
	}

	//MongoDB

	filter := bson.M{}

	var model []*Model.CasbinRoleModel

	//env

	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("casbinrolemodels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	cur, err := collection.Find(ctx, filter)

	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(ctx) {

		var elem Model.CasbinRoleModel

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
		"list":         model,
	})

}

func ApiAddCasbinRole(c *gin.Context) {

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
		"session_name": sessionName,
	})

}

func ApiDeleteCasbinRole(c *gin.Context) {

	var model Model.CasbinRoleModel

	//MongoDB
	DB_DATABASE := config.EnvFunc("DB_DATABASE")

	collection := config.MongoClient.Database(DB_DATABASE).Collection("casbinrolemodels")

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

	e := config.CasbinRole()

	e.RemovePolicy(
		model.RoleName,
		model.Path,
		model.Method,
	)

	_, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "err collections find one",
		})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"data": true})
}
