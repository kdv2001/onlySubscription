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
