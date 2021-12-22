# IMMUGORM
The [immudb](https://github.com/codenotary/immudb) gorm driver

## Quick Start

```go
import (
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
)

package main

import (
    "github.com/codenotary/immudb/pkg/client"
    immugorm "github.com/codenotary/immugorm"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type Product struct {
    Id     int `gorm:"primarykey"`
    Code   string
    Price  uint
    Amount uint
}

func main() {
    opts := client.DefaultOptions()

    opts.Username = "immudb"
    opts.Password = "immudb"
    opts.Database = "defaultdb"

    db, err := gorm.Open(immugorm.Open(opts, &immugorm.ImmuGormConfig{Verify: false}), &gorm.Config{
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
    // find product with integer primary key
    err = db.First(&product, 1).Error
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
    err = db.Delete(&product, 1).Error
    if err != nil {
        panic(err)
    }
}
```
