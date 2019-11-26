package handlers

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

func GetPro(w http.ResponseWriter, r *http.Request) {
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

	if money >= 10 {
		w.Write([]byte("<html><head></head><body>Hint: flag is spbctf{ka_ka_very_s3cur3_service}<br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><p>Haha, naebal</p></body></html>"))
	} else {
		w.Write([]byte("You have not enough money"))
	}
}
