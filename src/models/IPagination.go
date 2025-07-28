package models

// PaginationMetadata defines the metadata for paginated responses.
type PaginationMetadata struct {
	CurrentPage     int64 `json:"currentPage"`
	ItemsPerPage    int64 `json:"itemsPerPage"`
	TotalItems      int64 `json:"totalItems"`
	TotalPages      int64 `json:"totalPages"`
	HasNextPage     bool  `json:"hasNextPage"`
	HasPreviousPage bool  `json:"hasPreviousPage"`
}

// PaginatedResponse defines the structure for paginated API responses.
type PaginatedResponse struct {
	Data     any                `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}
