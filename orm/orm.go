package orm

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/bentranter/terrible/internal/reflectx"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/serenize/snaker"
)

// Connect creates a new instance of the ORM.
func Connect() *ORM {
	orm := &ORM{}

	conn, err := sqlx.Connect("postgres", "user=ben dbname=blog_development sslmode=disable")
	if err != nil {
		panic(err)
	}

	stmtCache := squirrel.NewStmtCacher(conn)
	builder := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		RunWith(stmtCache)

	orm.conn = conn
	orm.builder = builder
	return orm
}

// An ORM is an object relational mapper.
type ORM struct {
	conn    *sqlx.DB
	builder squirrel.StatementBuilderType
}

// Find gets an object by its primary key.
func (o *ORM) Find(model interface{}, id int64) error {
	t := time.Now()

	stmt, args, err := o.base().Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}
	defer o.log(t, stmt, args)

	return o.conn.Get(model, stmt, args...)
}

func (o *ORM) Save(model interface{}) error {
	t := time.Now()

	columns, values := o.columns(model)

	stmt, args, err := o.builder.Insert("articles").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return err
	}
	defer o.log(t, stmt, args)

	return o.conn.QueryRowx(stmt, args...).StructScan(model)
}

func (o *ORM) base() squirrel.SelectBuilder {
	return o.builder.Select("*").From("articles")
}

func (o *ORM) log(t time.Time, stmt string, args []interface{}) {
	color.New(color.FgHiCyan, color.Bold).Printf("Article Load (%s) ", time.Since(t))
	color.New(color.FgHiBlue, color.Bold).Print(stmt)
	fmt.Printf(" %v\n", args)
}

func (o *ORM) columns(model interface{}) ([]string, []interface{}) {
	rv := reflect.ValueOf(model)

	elem, err := reflectx.GetStructFromPtr(rv)
	if err != nil {
		panic(err)
	}

	numFields := elem.NumField()

	columns := make([]string, 0)
	values := make([]interface{}, 0)

	for i := 0; i < numFields; i++ {
		field := elem.Field(i)

		// If the field can't be set, any of the operations below will panic,
		// so we continue if that's the case.
		if !field.CanSet() {
			continue
		}

		name := elem.Type().Field(i).Name
		name = snaker.CamelToSnake(name)

		value := field.Interface()
		if value != reflect.Zero(field.Type()).Interface() {
			columns = append(columns, name)
			values = append(values, value)
		} else {
			switch name {
			case "id":
				continue

			case "created_at", "updated_at":
				columns = append(columns, name)
				values = append(values, time.Now())
			}

		}
	}

	return columns, values
}
