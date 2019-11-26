package handlers

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"io/ioutil"
	"kids-bank/flock"
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

	db, err := sql.Open("mysql", fmt.Sprintf("bank:%v@tcp(%v)/bank2",
		viper.GetString("MYSQL_PASSWORD"), viper.GetString("MYSQL_ADDR")))
	if err != nil {
		logrus.WithError(err).Errorf("Can't open database")
		http.Error(w, "Can't open database", http.StatusServiceUnavailable)
		return
	}
	defer db.Close()

	row := db.QueryRow("SELECT login, money FROM bank2.users WHERE INSTR(session, ?) > 0", session.Value)

	var login string
	var money int
	if err := row.Scan(&login, &money); err == sql.ErrNoRows {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		tmpl, err := template.New("share.html").ParseFiles("frontend/share.html")
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
		l, err := flock.NewPath(fmt.Sprintf("/tmp/%v.lock", session.Value))
		if err != nil {
			logrus.WithError(err).Errorf("Can't create flock")
			http.Error(w, "Can't create flock", http.StatusInternalServerError)
			return
		}
		l.Lock()
		defer l.Unlock()

		err = r.ParseForm()
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

		row := db.QueryRow("SELECT money, shared FROM bank2.users WHERE login=?", recepient)

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

		row = db.QueryRow("SELECT shared FROM bank2.users WHERE login=?", login)

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
		_, err = db.Exec("UPDATE bank2.users SET shared = 1, money = money+? WHERE login=?",
			amountInt, recepient)
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE bank2.users SET shared = 1, money = money-? WHERE login=?",
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
