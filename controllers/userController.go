package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"httpmon.com/first/database"
	helper "httpmon.com/first/helpers"
	"httpmon.com/first/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

//HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

//VerifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or passowrd is incorrect")
		check = false
	}

	return check, msg
}

//CreateUser is the api used to tget a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"name": user.Name})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the username"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this username already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Name, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

//Login is the api used to tget a single user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"name": user.Name}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Username or Password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Name, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		MonitorAllRequests(foundUser)
		c.JSON(http.StatusOK, foundUser.Token)

	}
}

func CreateUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userId, _ := c.Get("user_id")
		fmt.Println(userId)
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var url models.URL
		if err := c.BindJSON(&url); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(url.URL)
		if len(user.Urls) >= 20 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "You can Just add 20 Urls"})
			return
		}
		var result bool = false
		for _, x := range user.Urls {
			if x.URL == url.URL {
				result = true
				break
			}
		}
		if result {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this url already exists"})
			return
		}
		url.Failed = 0
		user.Urls = append(user.Urls, url)
		filter := bson.M{"user_id": userId}
		userCollection.ReplaceOne(ctx, filter, user)

		defer cancel()

		if err != nil {
			log.Panic(err)
			return
		}
		go RequestHTTP(user.User_id, url)
		c.JSON(http.StatusOK, user)
	}
}
func GetUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userId, _ := c.Get("user_id")
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user.Urls)
	}
}
func DeleteUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userId, _ := c.Get("user_id")
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var url models.URL
		if err := c.BindJSON(&url); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var result bool = false
		var index int = 0
		for i, x := range user.Urls {
			if x.URL == url.URL {
				result = true
				index = i
				break
			}
		}
		if !result {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this url not exists"})
			return
		}
		user.Urls = append(user.Urls[:index], user.Urls[index+1:]...)
		filter := bson.M{"user_id": userId}
		userCollection.ReplaceOne(ctx, filter, user)

		defer cancel()

		if err != nil {
			log.Panic(err)
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func MonitorAllRequests(user models.User) {
	for _, x := range user.Urls {
		fmt.Printf("url: %s  userId: %s\n", x.URL, user.User_id)
		go RequestHTTP(user.User_id, x)
	}
}

func RequestHTTP(userId string, url models.URL) {
	for true {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		var isURLExists bool = false
		var index int = 0
		for i, x := range user.Urls {
			if x.URL == url.URL {
				isURLExists = true
				index = i
				break
			}
		}
		if !isURLExists {
			break
		}
		resp, err := http.Get(url.URL)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("url: %s  statuscode: %d\n", url.URL, resp.StatusCode)
		var history models.History
		history.URL = url
		history.StatusCode = resp.StatusCode
		history.Requested_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if resp.StatusCode < 200 && resp.StatusCode > 299 {
			user.Urls[index].Failed++
			if user.Urls[index].Failed == user.Urls[index].Threshold {
				user.Alerts = append(user.Alerts, history)
				user.Urls[index].Failed = 0
			}

		} else {
			user.Urls[index].Succeed++
		}
		user.History = append(user.History, history)
		filter := bson.M{"user_id": userId}
		userCollection.ReplaceOne(ctx, filter, user)
		time.Sleep(10 * time.Second)
	}
}
func GetHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userId, _ := c.Get("user_id")
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user.History)
	}
}
func GetAlerts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userId, _ := c.Get("user_id")
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user.Alerts)
	}
}
