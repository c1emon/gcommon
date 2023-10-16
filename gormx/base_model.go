package gormx

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type BaseFields struct {
	Id         string    `json:"id" gorm:"column:id;type:char(25);primaryKey;<-:create"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;<-:create;autoUpdateTime"`
}

func (f *BaseFields) BeforeCreate(tx *gorm.DB) (err error) {
	f.Id = cuid.New()
	return nil
}

func CreateBaseField() BaseFields {
	return BaseFields{Id: cuid.New()}
}

func (f *BaseFields) GetId() string {
	return f.Id
}

func (f *BaseFields) GetCreatedTime() time.Time {
	return f.CreateTime
}

func (f *BaseFields) GetUpdatedTime() time.Time {
	return f.UpdateTime
}

func (f *BaseFields) SetUpdatedTime(t time.Time) {
	f.UpdateTime = t
}
