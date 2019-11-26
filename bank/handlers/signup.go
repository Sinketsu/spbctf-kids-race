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

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.New("signup.html").ParseFiles("frontend/signup.html")
		print(err)
		if err != nil {
			logrus.WithError(err).Errorf("Can't parse signup.html")
			http.Error(w, "Can't parse signup.html", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			logrus.WithError(err).Errorf("Can't execute signup.html")
			http.Error(w, "Can't execute signup.html", http.StatusInternalServerError)
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

		db, err := sql.Open("mysql", fmt.Sprintf("bank:%v@tcp(%v)/bank",
			viper.GetString("MYSQL_PASSWORD"), viper.GetString("MYSQL_ADDR")))
		if err != nil {
			logrus.WithError(err).Errorf("Can't open database")
			http.Error(w, "Can't open database", http.StatusServiceUnavailable)
			return
		}
		defer db.Close()

		row := db.QueryRow("select * from bank.users WHERE login=?", login)
		if err := row.Scan(nil); err != sql.ErrNoRows {
			http.Error(w, "User already registered", http.StatusAlreadyReported)
			return
		}

		session := uuid.NewV4().String()

		_, err = db.Exec("INSERT INTO bank.users (login, password, session) VALUES (?, ?, ?)",
			login, password, session)
		if err != nil {
			logrus.WithError(err).Errorf("Error while inserting new user")
			http.Error(w, "Can't exec query", http.StatusServiceUnavailable)
			return
		}

		authCookie := &http.Cookie{
			Name: "AUTH",
			Path: "/",
			Expires: time.Now().Add(10 * time.Hour),
			HttpOnly:true,
			Value: session,
		}

		http.SetCookie(w, authCookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
