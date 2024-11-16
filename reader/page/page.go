package page

import (
	"sync"

	"github.com/c1emon/gcommon/util"
)

// PageReadFn 用于读取`page`页的数据
type PageReadFn[T any] func(page int) (int, []T, error)

// PageReader 用于读取分页数据
type PageReader[T any] struct {
	fn PageReadFn[T]

	options
	once *sync.Once
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
			retTotal, retData, err := r.fn(page)
			if err == nil {
				r.once.Do(func() {
					total = retTotal
				})
				data = retData
			}
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

func NewPageReader[T any](fn PageReadFn[T], opts ...util.Option[options]) *PageReader[T] {

	opt := &options{
		offset:         0,
		nonZeroStartup: false,
		ignoreErr:      false,
	}
	for _, fn := range opts {
		fn.Apply(opt)
	}

	return &PageReader[T]{
		fn:      fn,
		options: *opt,
		once:    &sync.Once{},
	}
}
