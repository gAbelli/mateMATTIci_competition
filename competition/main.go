package main

func main() {
	SetupDB()
	r := SetupRouter()
	r.Run(":8080")
}
