package config

import (
	"os"

	"github.com/casbin/casbin/v2"
	mongodbadapter "github.com/casbin/mongodb-adapter/v2"
	"github.com/joho/godotenv"
)

func CasbinRole() *casbin.Enforcer {

	//env
	errenv := godotenv.Load()
	if errenv != nil {
		panic("Error loading .env file")
	}

	DB_USERNAME := os.Getenv("DB_USERNAME")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")

	//mongodb
	a, errcasbindb := mongodbadapter.NewAdapter("mongodb://" + DB_USERNAME + ":" + DB_PASSWORD + "@" + DB_HOST + ":" + DB_PORT)

	if errcasbindb != nil {
		panic("Failed to connect to database!")
	}
	e, errcasbin := casbin.NewEnforcer("config/Casbin_role_model.conf", a)

	if errcasbin != nil {
		panic("Failed to casbin!")
	}

	e.LoadPolicy()

	return e
}
