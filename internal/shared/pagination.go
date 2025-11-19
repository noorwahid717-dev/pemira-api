package shared

type PaginationParams struct {
	Page    int `json:"page" validate:"min=1"`
	PerPage int `json:"per_page" validate:"min=1,max=100"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func NewPaginationParams(page, perPage int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p PaginationParams) Limit() int {
	return p.PerPage
}

func NewPaginatedResponse(data interface{}, params PaginationParams, total int64) PaginatedResponse {
	totalPages := int(total) / params.PerPage
	if int(total)%params.PerPage > 0 {
		totalPages++
	}
	
	return PaginatedResponse{
		Data: data,
		Meta: PaginationMeta{
			CurrentPage: params.Page,
			PerPage:     params.PerPage,
			Total:       total,
			TotalPages:  totalPages,
		},
	}
}
