# IMMUGORM
The [immudb](https://github.com/codenotary/immudb) gorm driver

## Quick Start
Clone immudb repository, compile immudb and launch it:
```shell
git clone https://github.com/codenotary/immudb
make immudb
./immudb
```

Here an example
```go
package main

import (
    immugorm "github.com/codenotary/immugorm"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type Product struct {
    ID     int `gorm:"primarykey"`
    Code   string
    Price  uint
    Amount uint
}

func main() {
	db, err := gorm.Open(immugorm.Open("immudb://immudb:immudb@127.0.0.1:3322/defaultdb?sslmode=disable", &immugorm.ImmuGormConfig{Verify: false}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&Product{})
	if err != nil {
		panic(err)
	}
	// Create
	err = db.Create(&Product{Code: "D43", Price: 100, Amount: 500}).Error
	if err != nil {
		panic(err)
	}
	// Read
	var product Product
	// find just created one
	err = db.First(&product).Error
	if err != nil {
		panic(err)
	}
	// find product with code D42
	err = db.First(&product, "code = ?", "D43").Error
	if err != nil {
		panic(err)
	}
	// Update - update product's price to 200
	err = db.Model(&product).Update("Price", 888).Error
	if err != nil {
		panic(err)
	}

	// Update - update multiple fields
	err = db.Model(&product).Updates(Product{Price: 200, Code: "F42"}).Error
	if err != nil {
		panic(err)
	}

	err = db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"}).Error
	if err != nil {
		panic(err)
	}

	// Delete - delete product
	err = db.Delete(&product, product.ID).Error
	if err != nil {
		panic(err)
	}
}
```
## IMMUDB SPECIAL FEATURES

### TamperProof read
Immugorm is able to take benefits from immudb tamperproof capabilities.
It's possible to activate tamperproof read by setting  `Verify: true` when opening db.
The only difference is that verifications returns proofs needed to mathematically verify that the data was not tampered.
>Note that generating that proof has a slight performance impact.
>
```go
    db, err := gorm.Open(immugorm.Open(opts, &immugorm.ImmuGormConfig{Verify: true}), &gorm.Config{})
```
### Timetravel

Time travel allows reading data from SQL as if it was in some previous state.
> The state is indicated by transaction id, that is a monotonically increasing number that is assigned to each transaction by immudb
> Each operation is assigned a transaction id.
```go
db.Clauses(immugorm.BeforeTx(9)).Last(&entity, 1)
```

## Warnings

This is an experimental software. The API is not stable yet and may change without notice.
There are limitations:
* missing support related to altering or deleting already existent elements on schema. No drop table, index, alter table or column
* missing float type
* missing left join
* no support for composite primary key
* no support for polymorphism
* no support for foreign constraints
* is mandatory to have a primary key on tables
* no default values
* order is limit to one indexed column
* group by not supported
* no support for prepared statements
* not condition missing
* no transaction with savepoint
* no nested transactions
* having nor yet supported
