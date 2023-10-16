package errorx

import (
	"errors"

	"gorm.io/gorm"
)

var _ Parser = &GormParser{}

type GormParser struct {
}

func (p *GormParser) Parse(err error) ErrorX {
	if e, ok := err.(error); ok {
		switch {
		case errors.Is(e, gorm.ErrRecordNotFound):
			return ErrResourceNotFound
		}
	}
	return ErrUnknown
}

func (p *GormParser) Support(err error) bool {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, gorm.ErrNotImplemented):
		return true
	}
	return false
}

func NewGormParser() *GormParser {
	p := &GormParser{}
	Parsers.Add("gorm", p)
	return p
}
