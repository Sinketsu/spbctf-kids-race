package handlers

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"net/http"
)

type User struct {
	Login string
	Money int
}

func Index(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("AUTH")
	if err != nil {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
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

	row := db.QueryRow("SELECT login, money FROM bank.users WHERE session=?", session.Value)

	var login string
	var money int
	if err := row.Scan(&login, &money); err == sql.ErrNoRows {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("index.html").ParseFiles("frontend/index.html")
	print(err)
	if err != nil {
		logrus.WithError(err).Errorf("Can't parse index.html")
		http.Error(w, "Can't parse index.html", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &User{
		Login: login,
		Money: money,
	})
	if err != nil {
		logrus.WithError(err).Errorf("Can't execute index.html")
		http.Error(w, "Can't execute index.html", http.StatusInternalServerError)
		return
	}
}
