package main

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	r.POST("/submission", SubmissionHandler)
	r.GET("/leaderboard/:id", LeaderboardHandler)
	return r
}
