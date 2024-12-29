package test

import (
	"encoding/json"
	"testing"

	"github.com/c1emon/gcommon/util"
)

type baseParams struct {
	Name string
}

type baseBuilder struct {
	param baseParams
}

func (b *baseBuilder) SetName(name string) {
	b.param.Name = name
}

func NewBaseBuilder() baseBuilder {
	return baseBuilder{
		param: baseParams{},
	}
}

func (b *baseBuilder) Build() (baseParams, error) {
	return b.param, nil
}

type FullParams struct {
	*baseParams
	Ver string
}

type fullBuilde struct {
	baseBuilder
	param FullParams
}

func (b *fullBuilde) SetVer(ver string) {
	b.param.Ver = ver
}

func (b *fullBuilde) Build() (FullParams, error) {
	var zero FullParams
	baseParams, err := b.baseBuilder.Build()
	if err != nil {
		return zero, err
	}
	b.param.baseParams = &baseParams
	return b.param, nil
}

func NewFullBuilder() *fullBuilde {
	return &fullBuilde{
		baseBuilder: NewBaseBuilder(),
		param:       FullParams{},
	}
}

func TheApi(builder util.Builder[FullParams]) string {
	param, _ := builder.Build()
	v, _ := json.MarshalIndent(param, "", "  ")
	param.Name = "2.0"
	return string(v)
}

func Test_builder(t *testing.T) {
	builder := NewFullBuilder()

	builder.SetName("clemon")
	builder.SetVer("1.0")

	t.Log(TheApi(builder))

	t.Log(TheApi(builder))

}
