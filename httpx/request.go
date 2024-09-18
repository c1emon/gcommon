package httpx

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Order int

const (
	ASC  Order = 1
	DESC Order = -1
)

func (o Order) String() string {
	if o == ASC {
		return "ASC"
	} else {
		return "DESC"
	}
}

func ParserOrder(val any) Order {
	switch v := val.(type) {
	case int:
		if v < 0 {
			return DESC
		}
	case string:
		if strings.TrimSpace(strings.ToUpper(v)) == DESC.String() {
			return DESC
		}
	case bool:
		if !v {
			return DESC
		}
	}
	return ASC
}

type SortItem struct {
	Field string
	Order Order
}

func (i SortItem) Sql() string {
	return fmt.Sprintf("%s %s", i.Field, i.Order.String())
}

type Pageable interface {
	GetPageSize() int
	GetPageNum() int
	SetTotalItem(int64)
	SetTotalPage(int64)
	GetOffset() int
	GetSorts() []SortItem
}

var _ Pageable = &Pagination{}

type Pagination struct {
	Size      int        `json:"size,omitempty"`
	Page      int        `json:"page,omitempty"`
	TotalItem int64      `json:"total_item,omitempty"`
	TotalPage int64      `json:"total_page,omitempty"`
	Offset    int        `json:"-"`
	Sorts     []SortItem `json:"-"`
}

func (p *Pagination) GetPageSize() int {
	return p.Size
}

func (p *Pagination) GetPageNum() int {
	return p.Page
}

func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.Size
}

func (p *Pagination) SetTotalPage(total int64) {
	p.TotalPage = total
}

func (p *Pagination) SetTotalItem(total int64) {
	p.TotalItem = total
}

func (p *Pagination) GetSorts() []SortItem {
	return p.Sorts
}

func PaginationFromQuery(req *http.Request) *Pagination {
	paginationQuery := &Pagination{
		Size: 10, TotalItem: 0, Page: 1, Sorts: make([]SortItem, 0, 3),
	}

	queries := req.URL.Query()
	if s := queries.Get("size"); s != "" {
		if size, err := strconv.Atoi(s); err == nil {
			paginationQuery.Size = size
		}
	}
	if p := queries.Get("page"); p != "" {
		if page, err := strconv.Atoi(p); err == nil {
			paginationQuery.Page = page
		}
	}
	// if t := queries.Get("total"); t != "" {
	// 	if total, err := strconv.Atoi(t); err == nil {
	// 		paginationQuery.TotalItem = int64(total)
	// 	}
	// }
	if s := queries.Get("sort"); s != "" {
		reg := regexp.MustCompile(`(?i)(desc|asc)\\(([a-zA-Z]+[0-9a-zA-Z_]*)\\)`)
		for _, k := range reg.FindAllStringSubmatch(s, -1) {
			paginationQuery.Sorts = append(paginationQuery.Sorts, SortItem{Field: k[2], Order: ParserOrder(k[1])})
		}
	}

	return paginationQuery
}
