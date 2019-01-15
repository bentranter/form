# Form

A Go package for automatic form parsing of HTTP requests.

## API

### `form.Unmarshal(r *http.Request, v interface{}) error`

`form.Unmarshal` maps an HTTP request's form data to the given interface. The given interface must be a pointer to a struct.

_Example_

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/bentranter/terrible/form"
)

type User struct {
	ID             int64
	Name           string
	AccountBalance float64
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user := &User{}
		if err := form.Unmarshal(r, user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "%d has name %s and account balance %f", user.ID, user.Name, user.AccountBalance)
	})

	http.ListenAndServe(":3000", nil)
}

```
