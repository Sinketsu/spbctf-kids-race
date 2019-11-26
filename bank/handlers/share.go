package handlers

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func Share(w http.ResponseWriter, r *http.Request) {
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

	if r.Method == http.MethodGet {
		tmpl, err := template.New("share.html").ParseFiles("frontend/share.html")
		print(err)
		if err != nil {
			logrus.WithError(err).Errorf("Can't parse share.html")
			http.Error(w, "Can't parse share.html", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, &User{
			Login: login,
			Money: money,
		})
		if err != nil {
			logrus.WithError(err).Errorf("Can't execute share.html")
			http.Error(w, "Can't execute share.html", http.StatusInternalServerError)
			return
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			logrus.WithError(err).Errorf("Can't parse form")
			http.Error(w, "Can't parse form", http.StatusInternalServerError)
			return
		}

		recepient := r.Form.Get("login")
		amount := r.Form.Get("money")
		amountInt, err := strconv.Atoi(amount)
		if err != nil {
			http.Error(w, "Can't convert money to int", http.StatusBadRequest)
			return
		}

		if len(recepient) == 0 || len(amount) == 0 {
			http.Error(w, "`login` and `money` fields are required", http.StatusBadRequest)
			return
		}

		row := db.QueryRow("SELECT money, shared FROM bank.users WHERE login=?", recepient)

		var shared int
		var currentMoney int
		if err := row.Scan(&currentMoney, &shared); err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}

		if shared != 0 {
			http.Error(w, "User already received or sent money", http.StatusBadRequest)
			return
		}

		if amountInt < 0 {
			http.Error(w, "Money < 0", http.StatusBadRequest)
			return
		}

		if amountInt > money {
			http.Error(w, "You have not enough money", http.StatusBadRequest)
			return
		}

		row = db.QueryRow("SELECT shared FROM bank.users WHERE login=?", login)

		var selfShared int
		if err := row.Scan(&selfShared); err == sql.ErrNoRows {
			http.Error(w, "You not found", http.StatusBadRequest)
			return
		}

		if selfShared != 0 {
			http.Error(w, "You already receive or sent money", http.StatusBadRequest)
			return
		}

		delay := viper.GetDuration("DELAY")
		time.Sleep(delay)

		// If all right
		_, err = db.Exec("UPDATE bank.users SET shared = 1, money = money+? WHERE login=?",
			amountInt, recepient)
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE bank.users SET shared = 1, money = money-? WHERE login=?",
			amountInt, login)
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusInternalServerError)
			return
		}

		done, err := ioutil.ReadFile("frontend/done.html")
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusInternalServerError)
			return
		}
		w.Write(done)
	}
}
