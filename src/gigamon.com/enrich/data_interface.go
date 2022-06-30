package enrich

// DataInterface to store enrichment info
type DataInterface interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	SetBulk(pairs map[string]string) error
}
