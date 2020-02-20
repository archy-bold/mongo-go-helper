package pagination

// Result represents a pagination result
type Result struct {
	Items         []interface{}
	Total         int32
	PageSize      int32
	CurrentPage   int32
	NumberOfPages int32
}
