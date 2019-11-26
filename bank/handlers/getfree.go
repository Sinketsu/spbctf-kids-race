package handlers

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

func GetFree(w http.ResponseWriter, r *http.Request) {
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

	if money >= 0 {
		w.Write([]byte("Hint: try harder!"))
	} else {
		w.Write([]byte("You have not enough money"))
	}
}
