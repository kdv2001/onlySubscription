package primitives

// Pagination пагинация
type Pagination struct {
	Num    uint64
	Offset uint64
}

// IntervalFilter интервальный фильтр
type IntervalFilter[T comparable] struct {
	// From начальная точка интервала
	From T
	// To конечная точка интервала
	To T
}

// SortType тип сортировки
type SortType string

const (
	// Unknown неизвестный тип сортировки
	Unknown SortType = ""
	// Ascending сортировка по возрастанию
	Ascending SortType = "Ascending"
	// Descending сортировка по убыванию
	Descending SortType = "Descending"
)
