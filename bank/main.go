package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"kids-bank/handlers"
	"net/http"
)

func main() {

	viper.AutomaticEnv()

	http.HandleFunc("/", handlers.Index)
	http.HandleFunc("/signup", handlers.Signup)
	http.HandleFunc("/signin", handlers.Signin)
	http.HandleFunc("/share", handlers.Share)

	http.HandleFunc("/getfree", handlers.GetFree)
	http.HandleFunc("/getpro", handlers.GetPro)
	http.HandleFunc("/getanime", handlers.GetAnime)

	logrus.Fatal(http.ListenAndServe(viper.GetString("ADDR"), nil))
}
