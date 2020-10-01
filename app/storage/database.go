package storage

import (
	"app/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase() (Storage, error) {
	var database Database

	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		// TODO: maybe enable SSL
		// TODO: change to ticker and move to other func
		db, err = gorm.Open("postgres", "host=my-postgres-postgresql port=5432 user=postgres dbname=postgres password=postgres sslmode=disable")
		if err == nil {
			break
		}

		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	database.db = db
	db.AutoMigrate(&models.Run{})

	return &database, nil
}

func (d Database) AddRun(run *models.Run) error {
	return d.db.Create(run).Error
}

func (d Database) FindRun(query *models.Query) (*models.Run, error) {
	//var run models.Run
	var run models.Run
	err := d.db.Where(&models.Run{Query: models.Query{Hash: query.Hash}}).
		Take(&run).
		Error

	if err != nil {
		return nil, err
	}

	return &run, nil
}

// TODO: maybe add to interface
func (d Database) Close() error {
	return d.db.Close()
}
