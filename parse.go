package parse

import (
	"fmt"
	"encoding/json"
	"net/http"

  "github.com/jmcvetta/napping"
)

/**
 * Constants
 */

const PARSE_URL = "https://api.parse.com/1"

var APP_ID string
var API_KEY string
var session napping.Session

func InitAPI(appId, apiKey string) {
	APP_ID = appId
	API_KEY = apiKey

	header := http.Header{}
	header.Set("X-Parse-Application-Id", APP_ID)
	header.Set("X-Parse-REST-API-Key", API_KEY)

	session = napping.Session{Header: &header}
}

/**
 * Object type, base type for parse objects
 */

type Object struct {
	fields map[string]interface{}
}

// Constructor
func NewObject(fields map[string]interface{}) *Object {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	o := &Object{map[string]interface{}{}}
	o.SetFromMap(fields)
	return o
}

func (o *Object) SetFromMap(fields map[string]interface{}) {
	for key, value := range fields {
		if v, ok := value.(map[string]interface{}); ok == true {
			o.Set(key, NewObject(v))
		} else {
			o.Set(key, value)
		}
	}
}

// MarshalJSON, pass `fields` as own JSON representation
func (o *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.fields)
}

// Push a `value` to a slice under the `key` arg
func (o *Object) Push(key string, value interface{}) {
	sliceObj, ok := o.Get(key)
	if ok == false {
		sliceObj = make([]interface{}, 0, 10)
	}
	sliceObj = append(sliceObj.([]interface{}), value)
	o.Set(key, sliceObj)
}

func (o *Object) Set(key string, value interface{}) {
	o.fields[key] = value
}

func (o *Object) Get(key string) (value interface{}, ok bool) {
	value, ok = o.fields[key]
	return
}

// Typed methods

func (o *Object) GetString(key string) (value string, ok bool) {
	v, ok := o.Get(key)
	value = v.(string)
	return
}

func (o *Object) GetInt(key string) (value int, ok bool) {
	v, ok := o.Get(key)
	value = v.(int)
	return
}

func (o *Object) GetFloat(key string) (value float64, ok bool) {
	v, ok := o.Get(key)
	value = v.(float64)
	return
}

func (o *Object) GetObject(key string) (value *Object, ok bool) {
	v, ok := o.Get(key)
	value = v.(*Object)
	return
}

// Ensure that the value under a given `key` is an Object instance
func (o *Object) AssertObject(key string) (*Object, error) {
	v, ok := o.Get(key)
	if ok == false {
		v = NewObject(make(map[string]interface{}))
		o.Set(key, v)
		return v.(*Object), nil
	}
	if _, ok := v.(*Object); ok == false {
		return nil, fmt.Errorf("%s is already set with non object type", key)
	}
	return v.(*Object), nil
}

/**
 * ClassObject inherits from Object to add server methods
 */

type ClassObject struct {
	Object
	class string
}

// Constructors
func NewClassObject(class string) *ClassObject {
	co := &ClassObject{Object: *NewObject(nil), class: class}
	return co
}

func GetClassObject(class string, objectId string) (*ClassObject, error) {
	co := NewClassObject(class)
	co.Set("objectId", objectId)
	if err := co.get(); err != nil {
		return nil, err
	}
	return co, nil
}

// server methods
func (co *ClassObject) Save() error {
	if _, ok := co.Get("objectId"); ok == true {
		co.put()
	} else {
		co.post()
	}
	return nil
}

func (co *ClassObject) Delete() error {
	return nil
}

// private methods
func (co *ClassObject) get() error {
	objectId, ok := co.fields["objectId"]
	if ok == false {
		return fmt.Errorf("Object should have a `objectId` field to be gettable")
	}
	url := fmt.Sprintf("%s/classes/%s/%s", PARSE_URL, co.class, objectId)
	data := map[string]interface{}{}
	response, err := session.Get(url, nil, &data, nil)
	if err != nil {
		return err
	}
	if response.Status() == 200 {
		co.SetFromMap(data)
	} else {
		return fmt.Errorf("response status for Get requests should be 200, got %d", response.Status())
	}
	return nil
}

func (co *ClassObject) post() error {
	return nil
}

func (co *ClassObject) put() error {
	return nil
}

func (co *ClassObject) del() error {
	return nil
}
