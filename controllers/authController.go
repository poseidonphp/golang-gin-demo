package controllers

import (
	"context"
	"demo/initializers"
	"demo/models"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"net/http"
	"os"
	"time"
)

// OS Envs cannot go into the variables directly
var (
	oauth2Config = &oauth2.Config{
		RedirectURL: "http://localhost:8080/callback",
		Scopes:      []string{"read:user"},
		//Scopes:      []string{"openid", "profile", "email"},
		Endpoint: github.Endpoint,
		//Endpoint:     microsoft.AzureADEndpoint("your-tenant-id"),
	}
)

func LoginHandler(c *gin.Context) {
	oauth2Config.ClientID = os.Getenv("GITHUB_CLIENT_ID")
	oauth2Config.ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	// Generate the URL for AzureAD login
	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Println("Client ID: ", oauth2Config.ClientID)

	// Redirect user to AzureAD login page
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func CallbackHandler(c *gin.Context) {
	oauth2Config.ClientID = os.Getenv("GITHUB_CLIENT_ID")
	oauth2Config.ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	// Get the authorization code from the request
	code := c.Query("code")
	if code == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("no code in request"))
		return
	}

	// Exchange the authorization code for an access token
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Token received, you can now use it to call AzureAD protected APIs or get user info
	fmt.Println("Access Token:", token.AccessToken) //TODO: Do not print this to the console in production

	//TODO: Get the users info and avatar
	// TODO: Get or create user in DB

	// Store the token in a session or cookie as per your requirement
	// Example: c.SetCookie("token", token.AccessToken, 3600, "/", "localhost", false, true)
	user := map[string]string{
		"name":  "John Doe",
		"id":    "1234",
		"email": "test@test.com",
	}
	userObj, err := json.Marshal(user)
	session := sessions.Default(c)
	session.Set("token", token.AccessToken)
	session.Set("user", userObj)
	session.Save()
	c.Redirect(http.StatusTemporaryRedirect, "/")
	//c.String(http.StatusOK, "Login successful!")
}

/*
***************

	EVERYTHING BELOW THIS BLOCK IS AN
	 EXAMPLE FROM
	https://ututuv.medium.com/building-user-authentication-and-authorisation-api-in-go-using-gin-and-gorm-93dfe38e0612
*/
func Login(c *gin.Context) {

	var authInput models.AuthInput
	initializers.ConnectDB()

	if err := c.ShouldBindJSON(&authInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userFound models.User
	initializers.DB.Where("username=?", authInput.Username).Find(&userFound)

	if userFound.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFound.Password), []byte(authInput.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userFound.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := generateToken.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to generate token"})
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

func CreateUser(c *gin.Context) {

	var authInput models.AuthInput
	initializers.ConnectDB()

	if err := c.ShouldBindJSON(&authInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userFound models.User
	initializers.DB.Where("username=?", authInput.Username).Find(&userFound)

	if userFound.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already used"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(authInput.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Username: authInput.Username,
		Password: string(passwordHash),
	}

	initializers.DB.Create(&user)

	c.JSON(http.StatusOK, gin.H{"data": user})

}

func GetUserProfile(c *gin.Context) {

	user, _ := c.Get("currentUser")

	c.JSON(200, gin.H{
		"user": user,
	})
}
