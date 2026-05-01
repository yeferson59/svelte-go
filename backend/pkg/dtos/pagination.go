package dtos

type FilterPagination[I any, M any] struct {
	Data     I `json:"data"`
	MetaData M `json:"metaData"`
}
