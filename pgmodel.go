// Package pgmodel provides a generalized way of working with Postgres rows.
// Implementing the PGModel interface allows you to utilize the get, save, and
// delete functions without having to write repetitive query strings.
package pgmodel

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// PGModel interface types implement methods that describe their table.
type PGModel interface {

	// The model's primary key.
	PrimaryKey() string

	// The value of the model's primary key.
	PrimaryKeyValue() interface{}

	// The schema name of the model's table.
	SchemaName() string

	// The model's table name.
	TableName() string

	// The total number of columns in the table.
	ColumnCount() int

	// An array of non-primary key column names.
	NonPKColumns() []string

	// An array of non-primary key values in the same order as the columns defined
	// by NonPKColumns.
	NonPKValues() []interface{}

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
	ConvertSlice(c string) string
}

// MARK: Exported functions

// Get is identical to GetMany but QueryOne is called instead of Query on the
// transaction.
func Get(pm PGModel, t *pg.Tx, queryKey string, queryValue interface{}) (orm.Result, error) {
	return t.QueryOne(pm, createGetQuery(pm, queryKey, queryValue), queryValue)
}

// GetMany gets the entity defined by the slice of models in the given
// transaction by querying for the given queryKey and queryValue.
func GetMany(pm []PGModel, t *pg.Tx, queryKey string, queryValue interface{}) (orm.Result, error) {
	m := reflect.New(reflect.TypeOf(pm)).Elem().Interface().(PGModel)
	return t.Query(&pm, createGetQuery(m, queryKey, queryValue), queryValue)
}

// Save performs an upsert in the given transaction.
func Save(pm PGModel, t *pg.Tx) (orm.Result, error) {
	pkv := convertVariable(pm, pm.PrimaryKeyValue(), pm.PrimaryKey())
	npkv := convertVariables(pm)

	// Create total column/value slices
	v := append([]interface{}{pkv}, npkv...)

	// Create our inputs
	tv := append(v, npkv...)
	tv = append(tv, pkv)

	// Perform the query
	return t.Query(pm, createSaveQuery(pm), tv...)
}

// Delete deletes the model from the transaction.
func Delete(pm PGModel, t *pg.Tx) (orm.Result, error) {
	return t.Query(pm, createDeleteQuery(pm), pm.PrimaryKeyValue())
}

// MARK: Non-exported functions

// createGetQuery creates a get query from the given queryKey and queryValue.
func createGetQuery(pm PGModel, queryKey string, queryValue interface{}) string {
	// Get everything once
	sn := pm.SchemaName()
	tn := pm.TableName()

	// Create the query
	return fmt.Sprintf(
		`SELECT * FROM %s.%s
		WHERE %s = ?`,
		sn,
		tn,
		queryKey,
	)
}

// createSaveQuery creates a save query.
func createSaveQuery(pm PGModel) string {
	// Get everything once
	pk := pm.PrimaryKey()
	sn := pm.SchemaName()
	tn := pm.TableName()
	cc := pm.ColumnCount()
	npkc := pm.NonPKColumns()

	// Create total column/value slices
	c := append([]string{pk}, npkc...)

	// Create arrays to join
	var im, sm []string
	im = append(im, "?")
	for i := 0; i < cc-1; i++ {
		im = append(im, "?")
		sm = append(sm, fmt.Sprintf("%s = ?", npkc[i]))
	}

	// Create the query
	return fmt.Sprintf(
		`INSERT INTO %s.%s (%s) 
		VALUES (%s) 
		ON CONFLICT (%s) 
		DO UPDATE
		SET %s 
		WHERE %s.%s = ?`,
		sn,
		tn,
		strings.Join(c, ", "),
		strings.Join(im, ", "),
		pk,
		strings.Join(sm, ", "),
		tn,
		pk,
	)
}

// createDeleteQuery creates a delete query.
func createDeleteQuery(pm PGModel) string {
	// Get everything once
	pk := pm.PrimaryKey()
	sn := pm.SchemaName()
	tn := pm.TableName()

	// Create the query
	return fmt.Sprintf(
		`DELETE FROM %s.%s
		WHERE %s.%s = ?`,
		sn,
		tn,
		tn,
		pk,
	)
}

func convertVariables(pm PGModel) []interface{} {
	var cv []interface{}
	for i, u := range pm.NonPKValues() {
		cv = append(cv, convertVariable(pm, u, pm.NonPKColumns()[i]))
	}
	return cv
}

func convertVariable(pm PGModel, v interface{}, c string) interface{} {
	rt := reflect.TypeOf(v)
	switch rt.Kind() {
	case reflect.Slice:
		return pm.ConvertSlice(c)
	default:
		return v
	}
}
