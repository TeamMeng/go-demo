package main

import (
	// "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	// dsn := "host=localhost user=postgres password=postgres dbname=testdb port=5432 sslmode=disable"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db = db.Debug()

	db.AutoMigrate(&Product{})

	db.Create(&Product{Code: "D42", Price: 100})

	var product Product
	db.First(&product, 1)
	db.First(&product, "code = ?", "D42")

	db.Model(&product).Update("Price", 200)
	db.Model(&product).Updates(Product{Code: "F42", Price: 200})
	db.Delete(&product, 1)
}
