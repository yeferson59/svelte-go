package dtos

type FilterPagination[I any, M any] struct {
	Items    I `json:"items"`
	MetaData M `json:"metaData"`
}
