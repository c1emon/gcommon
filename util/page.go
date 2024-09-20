package util

type PageReadFn[T any] func(page int) (int, []T)

type PageReader[T any] struct {
	fn             PageReadFn[T]
	offset         int
	nonZeroStartup bool
}

func (r *PageReader[T]) Read() func(func([]T) bool) {
	page := r.offset
	if r.nonZeroStartup {
		page += 1
	}
	total := -1
	return func(yield func([]T) bool) {
		for {
			var data []T
			total, data = r.fn(page)
			if !yield(data) {
				return
			}
			page += 1
			if (!r.nonZeroStartup && page >= total) || (r.nonZeroStartup && page > total) {
				return
			}
		}
	}
}

func NewPageReader[T any](fn PageReadFn[T], offset int, nonZeroStartup bool) *PageReader[T] {
	return &PageReader[T]{
		fn:             fn,
		offset:         offset,
		nonZeroStartup: nonZeroStartup,
	}
}
