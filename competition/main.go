package main

import (
	"fmt"
	"os"
)

func main() {
	SetupDB()
	r := SetupRouter()
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8081"
	}
	r.Run(fmt.Sprintf(":%s", port))
}
