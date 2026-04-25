package vo

import (
	"net/http"
	"testing"
)

func TestPagination_GetOffset_clamps(t *testing.T) {
	p := &Pagination{Page: 0, Size: 0}
	if got := p.GetOffset(); got != 0 {
		t.Fatalf("GetOffset: got %d, want 0", got)
	}
	p = &Pagination{Page: 2, Size: 10}
	if got := p.GetOffset(); got != 10 {
		t.Fatalf("GetOffset: got %d, want 10", got)
	}
}

func TestPaginationFromQuery_sort(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/?sort=desc(name) asc(id)", nil)
	if err != nil {
		t.Fatal(err)
	}
	p := PaginationFromQuery(req)
	if len(p.Sorts) != 2 {
		t.Fatalf("sorts: len=%d, want 2", len(p.Sorts))
	}
	if p.Sorts[0].Field != "name" || p.Sorts[0].Order != DESC {
		t.Fatalf("sort[0]: %+v", p.Sorts[0])
	}
	if p.Sorts[1].Field != "id" || p.Sorts[1].Order != ASC {
		t.Fatalf("sort[1]: %+v", p.Sorts[1])
	}
}
