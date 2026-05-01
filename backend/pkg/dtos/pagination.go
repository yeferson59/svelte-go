package dtos

type FilterPagination[I any, M any] struct {
	Data     I
	MetaData M
}
