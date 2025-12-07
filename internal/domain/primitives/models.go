package primitives

type Pagination struct {
	Num    uint64
	Offset uint64
}

type RequestList struct {
	Pagination Pagination
	// также сюда можно добавить фильтры
	// и сортировку
}

type IntervalFilter[T comparable] struct {
	From T
	To   T
}

type SortType string

const (
	Unknown    SortType = ""
	Ascending  SortType = "Ascending"
	Descending SortType = "Descending"
)
