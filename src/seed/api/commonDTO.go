package api

type MappingPagedDTO[T any] struct {
	Number        int  `json:"number"`
	Data          []T  `json:"data"`
	TotalElements int  `json:"totalElements"`
	TotalPages    int  `json:"totalPages"`
	Last          bool `json:"last"`
}

type CorePagedDTO[T any] struct {
	Number        int  `json:"number"`
	Content       []T  `json:"content"`
	TotalElements int  `json:"totalElements"`
	TotalPages    int  `json:"totalPages"`
	Last          bool `json:"last"`
}
