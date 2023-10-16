package gormx

import "context"

type BaseRepository[T any] interface {
	CreateOne(context.Context, *T) error
	GetOneById(context.Context, string) (*T, error)
	UpdateOneById(context.Context, string, *T) error
	DeleteOneById(context.Context, string) error
	InitDB() error
}
