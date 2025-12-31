package pagination

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	DefaultLimit  = 20
	MaxLimit      = 100
	DefaultOffset = 0
)

// Params holds pagination parameters
type Params struct {
	Limit  int
	Offset int
	Page   int // 1-indexed page number (alternative to offset)
}

// Response holds pagination metadata
type Response struct {
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	Page       int         `json:"page"`
	TotalPages int         `json:"totalPages"`
	HasNext    bool        `json:"hasNext"`
	HasPrev    bool        `json:"hasPrev"`
	Data       interface{} `json:"data"`
}

// ParseParams extracts pagination parameters from HTTP request
func ParseParams(r *http.Request) Params {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	pageStr := r.URL.Query().Get("page")

	limit := DefaultLimit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > MaxLimit {
				limit = MaxLimit
			}
		}
	}

	offset := DefaultOffset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	page := 0
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
			offset = (p - 1) * limit
		}
	}

	return Params{
		Limit:  limit,
		Offset: offset,
		Page:   page,
	}
}

// NewResponse creates a pagination response
func NewResponse(data interface{}, total int64, limit, offset int) Response {
	page := (offset / limit) + 1
	if offset == 0 && limit == 0 {
		page = 1
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return Response{
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		Page:       page,
		TotalPages: totalPages,
		HasNext:    offset+limit < int(total),
		HasPrev:    offset > 0,
		Data:       data,
	}
}

// ValidateParams validates pagination parameters
func ValidateParams(limit, offset int) error {
	if limit < 1 {
		return fmt.Errorf("limit must be greater than 0")
	}
	if limit > MaxLimit {
		return fmt.Errorf("limit cannot exceed %d", MaxLimit)
	}
	if offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}
	return nil
}

// CalculateOffset calculates offset from page number
func CalculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}

// CalculatePage calculates page number from offset
func CalculatePage(offset, limit int) int {
	if limit == 0 {
		return 1
	}
	return (offset / limit) + 1
}
