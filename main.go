package main

import (
	"iybe/routes"
	"iybe/utils"
	"iybe/controllers"
	"iybe/models"

	"fmt"
	"net/http"
	"github.com/rs/cors"
)

func main() {
	q := controllers.QWrapper{Queue: utils.NewQ()}
	routes := routes.Routes(&q)

	port := "8001"
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:3000", "http://localhost:3001"},
		AllowCredentials: true,
		AllowedMethods:   []string{"POST", "GET", "OPTIONS", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Origin", "X-Requested-With",
			                         "Content-Length", "Accept-Encoding", "Cache-Control",
															 "Authorization"},
		Debug: true,
	})

	handler := c.Handler(routes)

	ok := models.Seed()
	if !ok {
		panic("db not seeded")
	}
	fmt.Println("db seeded")
	err := http.ListenAndServe(":"+port, handler) //Launch the server
	if err != nil {
		panic(err)
	}
}
