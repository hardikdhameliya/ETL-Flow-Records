package upload

//DataStore interface for flow records
type DataStore interface {
	Set(m map[string][]interface{}) error
}
