package form_test

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/bentranter/terrible/form"
)

type S struct {
	Key        string
	Field      string
	unexported string
	Num        int
	F          float64
	B          bool
}

func TestFormUnmarshal(t *testing.T) {
	t.Parallel()

	expected := &S{Key: "value", Field: "Hello, world!", Num: 123, F: 0.1, B: true}
	body := strings.NewReader(url.Values{
		"Key":   []string{expected.Key},
		"Field": []string{expected.Field},
		"Num":   []string{strconv.Itoa(expected.Num)},
		"F":     []string{strconv.FormatFloat(expected.F, 'f', -1, 64)},
		"B":     []string{strconv.FormatBool(expected.B)},
	}.Encode())

	req, err := http.NewRequest(http.MethodPost, "/", body)
	if err != nil {
		t.Fatalf("error in test setup: failed to create request: %+v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	v := &S{}
	if err := form.Unmarshal(req, v); err != nil {
		t.Fatalf("error unmarshaling: %+v", err)
	}

	if !reflect.DeepEqual(v, expected) {
		t.Fatalf("not equal: expected %+v but got %+v", expected, v)
	}
}

func TestFormFor(t *testing.T) {
	t.Parallel()

	v := &S{
		Key: "value",
	}

	html := form.For(v, &form.ForOpts{
		Method: http.MethodPost,
		Action: "/users/create",
	})

	fmt.Println(html)
	t.Fail()
}
