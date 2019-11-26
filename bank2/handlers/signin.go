package handlers

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"net/http"
	"time"
)

func Signin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.New("signin.html").ParseFiles("frontend/signin.html")
		print(err)
		if err != nil {
			logrus.WithError(err).Errorf("Can't parse signin.html")
			http.Error(w, "Can't parse signin.html", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			logrus.WithError(err).Errorf("Can't execute signin.html")
			http.Error(w, "Can't execute signin.html", http.StatusInternalServerError)
			return
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			logrus.WithError(err).Errorf("Can't parse form")
			http.Error(w, "Can't parse form", http.StatusInternalServerError)
			return
		}

		login := r.Form.Get("login")
		password := r.Form.Get("password")

		if len(login) == 0 || len(password) == 0 {
			http.Error(w, "`login` and `password` fields are required", http.StatusBadRequest)
			return
		}

		db, err := sql.Open("mysql", fmt.Sprintf("bank:%v@tcp(%v)/bank2",
			viper.GetString("MYSQL_PASSWORD"), viper.GetString("MYSQL_ADDR")))
		if err != nil {
			logrus.WithError(err).Errorf("Can't open database")
			http.Error(w, "Can't open database", http.StatusServiceUnavailable)
			return
		}
		defer db.Close()

		row := db.QueryRow("SELECT session FROM bank2.users WHERE login=? AND password=?", login, password)
		var session string
		if err := row.Scan(&session); err == sql.ErrNoRows {
			http.Error(w, "User not registered", http.StatusConflict)
			return
		}

		newSession := uuid.NewV4().String()

		_, err = db.Exec("UPDATE users SET session = CONCAT(session, ';', ?) WHERE login=? AND password=?",
			newSession, login, password)
		if err != nil {
			logrus.WithError(err).Errorf("Can't update session")
			http.Error(w, "Can't update session", http.StatusServiceUnavailable)
			return
		}

		authCookie := &http.Cookie{
			Name: "AUTH",
			Path: "/",
			Expires: time.Now().Add(10 * time.Hour),
			HttpOnly:true,
			Value: newSession,
		}

		http.SetCookie(w, authCookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
