package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tjarratt/babble"
)

type Company struct {
	TableName struct{} `sql:"companies"`
	ID        int
	Name      string
	CEO       *Person
	CEOid     int   `sql:"ceo_id"`
}

type Person struct {
	TableName struct{} `sql:"people"`
	ID        int
	Name      string
}

const (
	wordsCount = 5
)

var babbler = babble.NewBabbler() // random phrases generator
var connectionString string
var db *pg.DB

func TestMain(m *testing.M) {
	flag.StringVar(&connectionString, "postgres", "postgres://postgres:postgres@localhost:5439/companies?sslmode=disable", "connection string for postgres")
	flag.Parse()
	babbler.Separator = " "
	babbler.Count = wordsCount
	var err error
	db, err = connectToPostgresTimeout(connectionString, 10*time.Second, time.Second)
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
	/*if err := createSchema(db); err != nil {
		logrus.Fatalf("%+v", err)
	}*/
	os.Exit(m.Run())
}

func TestHasOne(t *testing.T) {
	as := assert.New(t)
	ceo := &Person{Name: "ceo 1" + babbler.Babble()}
	if !as.NoError(db.Insert(ceo)) {
		return
	}
	com := &Company{
		Name:  babbler.Babble(),
		CEO:   ceo,
		CEOid: ceo.ID,
	}
	if !as.NoError(db.Insert(com)) {
		return
	}
	var compSelect Company
	if !as.NoError(db.Model(&compSelect).Column("company.*").Relation("CEO").Where("company.name = ?", com.Name).Select()) {
		return
	}
	if !as.NotZero(compSelect.ID) {
		return
	}
	if !as.NotNil(compSelect.CEO) {
		return
	}

	if !as.Equal(ceo.Name, compSelect.CEO.Name) {
		return
	}
}

func connectToPostgresTimeout(connectionString string, timeout, retry time.Duration) (*pg.DB, error) {
	var (
		connectionError error
		db              *pg.DB
	)
	connected := make(chan bool)
	go func() {
		for {
			db, connectionError = connectToPostgres(connectionString)
			if connectionError != nil {
				time.Sleep(retry)
				continue
			}
			connected <- true
			break
		}
	}()
	select {
	case <-time.After(timeout):
		err := errors.Wrapf(connectionError, "timeout %s connecting to db", timeout)
		return nil, err
	case <-connected:
	}
	return db, nil
}

func connectToPostgres(connectionString string) (*pg.DB, error) {
	opt, err := pg.ParseURL(connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to postgres with connection string: "+connectionString)
	}

	db := pg.Connect(opt)
	_, err = db.Exec("SELECT 1")
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}

	return db, nil
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{(*Company)(nil), (*Person)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
			//Temp: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
