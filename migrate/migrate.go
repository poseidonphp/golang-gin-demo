package main

import (
	"demo/initializers"
	"demo/models"
)

func init() {
	initializers.LoadEnvs()
	initializers.ConnectDB()

}

func main() {

	initializers.DB.AutoMigrate(&models.User{})
}
