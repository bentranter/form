package main

import (
	"fmt"
	"net/http"

	"github.com/bentranter/form"
)

type User struct {
	ID             int64
	Name           string
	AccountBalance float64
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			user := &User{
				Name: "Ben",
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(form.For(user)))

		case http.MethodPost:
			user := &User{}
			if err := form.Unmarshal(r, user); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Fprintf(w, "%d has name %s and account balance %f", user.ID, user.Name, user.AccountBalance)

		default:
			// ok
		}
	})

	http.ListenAndServe(":3000", nil)
}
