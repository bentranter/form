package orm_test

import (
	"testing"
	"time"
)

type article struct {
	ID        int64
	Title     string
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func TestORM(t *testing.T) {

}
