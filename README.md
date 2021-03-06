# package pgmodel

[![Go Reference](https://pkg.go.dev/badge/github.com/colinc86/pgmodel.svg)](https://pkg.go.dev/github.com/colinc86/pgmodel)

Package pgmodel provides a generalized way of working with Postgres rows. Implementing the PGModel interface allows you to utilize the get, save, and delete functions without having to write repetitive query strings.

Turn this:

```go
func (b *Bar) save(tx *pg.Tx) (orm.Result, error) {
  return tx.Query(b,
    `INSERT INTO foo.bars (id, name, date, value) 
    VALUES (?, ?, ?, ?) 
    ON CONFLICT (id) 
    DO UPDATE
    SET name = ?, date = ?, value = ? 
    WHERE bar.id = ?`, b.ID, b.Name, b.Date, b.Value, b.Name, b.Date, b.Value, b.ID)
}
```

In to this:

```go
func (b *Bar) save(tx *pg.Tx) (orm.Result, error) {
  return pgmodel.Save(b, tx)
}
```

## Install

```bash
$ go get github.com/colinc86/pgmodel
```

## Example

To utilize the `Get`, `GetMany`, `Save` and `Delete` functions, you must implement the `PGModel` interface.

```go
package example

import (
  "fmt"

  "github.com/go-pg/pg/v10"
  "github.com/go-pg/pg/v10/orm"
  "github.com/google/uuid"
)

type Bar struct {
  tableName struct{}  `pg:"bars"`
  ID        uuid.UUID `pg:"id,pk"`
  Name      string    `pg:"name"`
  Value     int       `pg:"value"`
  Values    []float64 `pg:"values,array"`
}

// PGModel interface methods

// PrimaryKey returns the model's primary key.
func (b Bar) PrimaryKey() string {
  return "id"
}

// PrimaryKeyValue returns the value of the model's primary key.
func (b Bar) PrimaryKeyValue() interface{} {
  return b.ID
}

// SchemaName returns the name of the model's schema.
func (b Bar) SchemaName() string {
  return "foo"
}

// TableName returns the name of the model's table.
func (b Bar) TableName() string {
  return "bars"
}

// ColumnCount returns the total number of columns in the table.
func (b Bar) ColumnCount() int {
  return 3
}

// NonPKColumns returns an array of non-primary key column names.
func (b Bar) NonPKColumns() []string {
  return []string{"name", "value"}
}

// NonPKValues returns an array of non-primary key values.
func (b Bar) NonPKValues() []interface{} {
  return []interface{}{b.Name, b.Value}
}

// Converts the model's slice from column, c, to a string value.
//
// This method's return value should be a stirng appropriate for a query. I.e.
// if the value in column c is the slice of floats, [0.1, 0.2, 0.3], then the
// method should return the string
//
//     {0.1, 0.2, 0.3}
//
// whereas if the value in c is the byte slice [0xA3, 0xA4, 0xA5], then the
// method should return
//
//     a3a4a5
func (b Bar) ConvertSlice(c string) string {
  var convertedValues []string
  if c == "values" {
    for _, v := range b.values {
      convertedValues = append(convertedValues, fmt.Sprintf("%f", v))
    }
  }

  return "{" + strings.Join(convertedValues, ",") + "}"
}
```

After implementing the `PGModel` interface, it is possible to write functions/methods such as...

```go
// Save saves the bar.
func (b *Bar) Save(tx *pg.Tx) (orm.Result, error) {
  return pgmodel.Save(b, tx)
}

// Delete deletes the bar.
func (b *Bar) Delete(tx *pg.Tx) (orm.Result, error) {
  return pgmodel.Delete(b, tx)
}

// GetBar gets a bar from the db.
func GetBar(tx *pg.Tx, queryKey string, queryValue interface{}) (*Bar, error) {
  b := new(Bar)
  _, err := pgmodel.Get(b, tx, queryKey, queryValue)
  if err != nil {
    return nil, err
  }
  return b, nil
}

// GetBar gets bars from the db.
func GetBars(tx *pg.Tx, queryKey string, queryValue interface{}) ([]*Bar, error) {
  var b []*Bar
  _, err := pgmodel.GetMany(b, tx, queryKey, queryValue)
  if err != nil {
    return nil, err
  }
  return b, nil
}
```
