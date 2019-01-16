package orm

import (
	"errors"
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

// All returns all models.
func (o *ORM) All(models interface{}) error {
	t := time.Now()

	stmt, args, err := o.base().ToSql()
	if err != nil {
		return err
	}
	defer o.log(t, stmt, args)

	return o.conn.Select(models, stmt, args...)
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

// Save inserts the model into the database. Any fields set by the database
// in the RETURNING clauses will be set on the model.
func (o *ORM) Save(model interface{}) error {
	t := time.Now()
	_, clauses := o.clauses(model, true)

	stmt, args, err := o.builder.Insert("articles").
		SetMap(clauses).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return err
	}
	defer o.log(t, stmt, args)

	return o.conn.QueryRowx(stmt, args...).StructScan(model)
}

// Update updates the model by its primary key. It will set the "updated_at"
// field to the current time.
func (o *ORM) Update(model interface{}) error {
	t := time.Now()
	id, clauses := o.clauses(model, false)

	stmt, args, err := o.builder.Update("articles").
		SetMap(clauses).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return err
	}
	defer o.log(t, stmt, args)

	return o.conn.QueryRowx(stmt, args...).StructScan(model)
}

// Destroy deletes a model by its primary key.
func (o *ORM) Destroy(model interface{}) error {
	t := time.Now()

	id, err := o.id(model)
	if err != nil {
		return err
	}

	stmt, args, err := o.builder.Delete("articles").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	defer o.log(t, stmt, args)

	_, err = o.conn.Exec(stmt, args...)
	return err
}

func (o *ORM) base() squirrel.SelectBuilder {
	return o.builder.Select("*").From("articles")
}

func (o *ORM) log(t time.Time, stmt string, args []interface{}) {
	color.New(color.FgHiCyan, color.Bold).Printf("Article Load (%s) ", time.Since(t))
	color.New(color.FgHiBlue, color.Bold).Print(stmt)
	fmt.Printf(" %v\n", args)
}

func (o *ORM) clauses(model interface{}, saving bool) (int64, map[string]interface{}) {
	rv := reflect.ValueOf(model)

	elem, err := reflectx.GetStructFromPtr(rv)
	if err != nil {
		panic(err)
	}

	numFields := elem.NumField()
	now := time.Now()
	clauses := make(map[string]interface{})
	id := int64(0)

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
			// Within this block, we handle fields that have values.
			//
			// We want different behaviour depending on the name of the field,
			// as some fields will be treated specially.
			switch name {

			// Don't update the ID if we're not saving. IDs shouldn't be
			// set in an "UPDATE" clause! Also, set the ID so we can return it
			// for the query engine to use as a primary key.
			case "id":
				if !saving {
					id = field.Int()
					continue
				} else {
					clauses[name] = value
				}

			// Don't update the "created_at" field if we're updating.
			case "created_at":
				if !saving {
					continue
				}

			// Update the "updated_at" field if we're updating, even if it has
			// a value.
			case "updated_at":
				if !saving {
					clauses[name] = now
				}

			default:
				clauses[name] = value
			}

		} else {
			// Within this block, we handle fields that DO NOT have values.
			switch name {
			// If the ID field is present, we want the database to set it,
			// so we do nothing here.
			case "id":
				continue

			// If we're saving (instead of updating) we want to set the
			// "created_at" field.
			case "created_at":
				if saving {
					clauses[name] = now
				}

			// If the "updated_at" field is null, we assume we always
			// want to update it.
			case "updated_at":
				clauses[name] = now
			}

		}
	}

	return id, clauses
}

func (o *ORM) id(model interface{}) (int64, error) {
	rv := reflect.ValueOf(model)

	elem, err := reflectx.GetStructFromPtr(rv)
	if err != nil {
		panic(err)
	}

	field := elem.FieldByName("ID")
	if field == reflect.Zero(field.Type()).Interface() {
		return 0, errors.New("id is nil")
	}
	return field.Int(), nil
}
