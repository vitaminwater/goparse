package address

import (
	"git.ccsas.biz/zilia_parse"
)

type Address struct {
	parse.ClassObject
}

func NewAddress() Address {
	a := Address{parse.NewClassObject("Address")}
	return a
}

func (a Address) SetName(name string) {
	a.Set("name", name)
}

func (a Address) SetDescription(description string) {
	a.Set("description", description)
}

func (a Address) PushImage(image string) {
	a.Push("images", image)
}

func (a Address) SetLocInfo(key string, value interface{}) {
	o, _ := a.AssertObject("locInfos")
	o.Set(key, value)
}

func (a Address) PushMetro(station string) {
	o, _ := a.AssertObject("locInfos")
	o.Push("metro", station)
}

func (a Address) SetDetail(key string, value interface{}) {
	o, _ := a.AssertObject("details")
	o.Set(key, value)
}
