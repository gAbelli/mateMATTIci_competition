package main

import (
	"context"
	"log"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"

	"google.golang.org/api/option"
)

func main() {
	opt := option.WithCredentialsFile(
		"credentials.json",
	)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	// gin.SetMode(gin.ReleaseMode)
	r.GET("/", func(c *gin.Context) {
		idToken := c.GetHeader("Authorization")
		ok := strings.HasPrefix(idToken, "Bearer ")
		if !ok {
			c.AbortWithStatus(403)
			return
		}
		idToken = idToken[7:]

		client, err := app.Auth(c)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		token, err := client.VerifyIDToken(c, idToken)
		if err != nil {
			c.AbortWithStatus(403)
			return
		}

		c.JSON(200, gin.H{
			"message": "OK",
		})
		log.Printf("Verified ID token: %v\n", token)
	})
	r.Run(":8081")
}
