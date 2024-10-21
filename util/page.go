package util

// PageReadFn 用于读取`page`页的数据
type PageReadFn[T any] func(page int) (int, []T, error)

// PageReader 用于读取分页数据
type PageReader[T any] struct {
	fn             PageReadFn[T]
	offset         int
	nonZeroStartup bool
	ignoreErr      bool
}

func (r *PageReader[T]) Read() func(func([]T, error) bool) {
	page := r.offset
	if r.nonZeroStartup {
		page += 1
	}
	total := -1
	return func(yield func([]T, error) bool) {
		for {
			var data []T
			var err error
			total, data, err = r.fn(page)
			if !yield(data, err) {
				return
			}
			if err != nil && !r.ignoreErr {
				return
			}
			page += 1
			if (!r.nonZeroStartup && page >= total) || (r.nonZeroStartup && page > total) {
				return
			}
		}
	}
}

func NewPageReader[T any](fn PageReadFn[T], offset int, nonZeroStartup bool, ignoreErr bool) *PageReader[T] {
	return &PageReader[T]{
		fn:             fn,
		offset:         offset,
		nonZeroStartup: nonZeroStartup,
		ignoreErr:      ignoreErr,
	}
}
