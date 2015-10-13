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
var REST_API_KEY string
var session napping.Session

func InitAPI(appId, apiKey string) {
	APP_ID = appId
	REST_API_KEY = apiKey

	header := http.Header{}
	header.Set("X-Parse-Application-Id", APP_ID)
	header.Set("X-Parse-REST-API-Key", REST_API_KEY)

	session = napping.Session{Header: &header}
}

/**
 * Object type, base type for parse objects
 */

type Object struct {
	changedKeys []string
	fields map[string]interface{}
}

// Constructor
func NewObject() *Object {
	changedKeys := []string{}
	fields := map[string]interface{}{}

	return &Object{changedKeys, fields}
}

func (o *Object) setKeyChanged(key string) {
	if inArray(o.changedKeys, key) == false {
		o.changedKeys = append(o.changedKeys, key)
	}
}

func (o *Object) didKeyChange(key string) bool {
	if obj, ok := o.GetObject(key); ok == true {
		return obj.hasChanges()
	}
	return inArray(o.changedKeys, key)
}

func (o *Object) hasChanges() bool {
	return len(o.changedKeys) != 0
}

func (o *Object) clearChangedKeys() {
	for _, v := range o.fields {
		if obj, ok := v.(*Object); ok == true {
			obj.clearChangedKeys()
		}
	}
	o.changedKeys = []string{}
}

func (o *Object) SetFromMap(fields map[string]interface{}) {
	for key, value := range fields {
		if v, ok := value.(map[string]interface{}); ok == true {
			nested := NewObject()
			nested.SetFromMap(v)
			o.Set(key, nested)
		} else {
			o.Set(key, value)
		}
	}
	o.clearChangedKeys()
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
	o.setKeyChanged(key)
	o.fields[key] = value
}

func (o *Object) SetNested(path []string, value interface{}) error {
	c := path[0]
	path = path[1:]
	if len(path) > 0 {
		co, err := o.AssertObject(c)
		if err != nil {
			return err
		}
		return co.SetNested(path, value)
	} else {
		o.Set(c, value)
	}
	return nil
}

func (o *Object) Get(key string) (value interface{}, ok bool) {
	value, ok = o.fields[key]
	return
}

func (o *Object) Remove(key string) {
	delete(o.fields, key)
	o.setKeyChanged(key)
}

// Typed methods
// TODO check type inference
func (o *Object) GetString(key string) (value string, ok bool) {
	v, ok := o.Get(key)
	if ok == false {
		return
	}
	value, ok = v.(string)
	return
}

func (o *Object) GetInt(key string) (value int, ok bool) {
	v, ok := o.Get(key)
	if ok == false {
		return
	}
	value, ok = v.(int)
	return
}

func (o *Object) GetFloat(key string) (value float64, ok bool) {
	v, ok := o.Get(key)
	if ok == false {
		return
	}
	value, ok = v.(float64)
	return
}

func (o *Object) GetObject(key string) (value *Object, ok bool) {
	v, ok := o.Get(key)
	if ok == false {
		return
	}
	value, ok = v.(*Object)
	return
}

func (o *Object) GetArray(key string) (value []interface{}, ok bool) {
	v, ok := o.Get(key)
	if ok == false {
		return
	}
	value, ok = v.([]interface{})
	return
}

// Ensure that the value under a given `key` is an Object instance
func (o *Object) AssertObject(key string) (*Object, error) {
	v, ok := o.Get(key)
	if ok == false {
		v = NewObject()
		o.Set(key, v)
		return v.(*Object), nil
	}
	if _, ok := v.(*Object); ok == false {
		return nil, fmt.Errorf("%s is already set with non object type", key)
	}
	return v.(*Object), nil
}

// toMap
func (o *Object) toMap(filter func(*Object, string) bool) map[string]interface{} {
	fields := map[string]interface{}{}
	for k, v := range o.fields {
		if filter(o, k) {
			if obj, ok := v.(*Object); ok == true {
				fields[k] = obj.toMap(filter)
			} else {
				fields[k] = v
			}
		}
	}
	return fields
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
	co := &ClassObject{Object: *NewObject(), class: class}
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

// Create a copy of this object, but removes Parse read only fields
func (co *ClassObject) postMap() map[string]interface{} {
	excluded := []string{"parsemap_identifier", "updatedAt", "createdAt", "objectId"}
	fields := co.toMap(func(o *Object, key string) bool {
		return !inArray(excluded, key)
	})
	return fields
}

func (co *ClassObject) putMap() map[string]interface{} {
	excluded := []string{"parsemap_identifier", "updatedAt", "createdAt", "objectId"}
	fields := co.toMap(func(o *Object, key string) bool {
		return !inArray(excluded, key) && ((o == &co.Object && o.didKeyChange(key)) || (o != &co.Object && o.hasChanges()))
	})
	return fields
}

// server methods
func (co *ClassObject) Save() error {
	if _, ok := co.Get("objectId"); ok == true {
		return co.put()
	} else {
		return co.post()
	}
}

func (co *ClassObject) Delete() error {
	return co.del()
}

// private methods
func (co *ClassObject) get() error {
	objectId, ok := co.Get("objectId")
	if ok == false {
		return fmt.Errorf("Object should have a `objectId` field to be gettable")
	}
	url := fmt.Sprintf("%s/classes/%s/%s", PARSE_URL, co.class, objectId)

	errorData := map[string]interface{}{}
	
	responseData := map[string]interface{}{}
	response, err := session.Get(url, nil, &responseData, errorData)
	if err != nil {
		return err
	}
	if statusOk(response.Status()) {
		co.SetFromMap(responseData)
	} else {
		errorJson, err := json.Marshal(errorData)
		if err != nil {
			return err
		}
		return fmt.Errorf("response status for Get requests should be 200, got %d\nerror: %s", response.Status(), errorJson)
	}
	return nil
}

func (co *ClassObject) post() error {
	url := fmt.Sprintf("%s/classes/%s", PARSE_URL, co.class)
	
	errorData := map[string]interface{}{}
	
	requestObj := co.postMap()
	responseData := map[string]interface{}{}
	response, err := session.Post(url, requestObj, &responseData, &errorData)
	if err != nil {
		return err
	}
	if !statusOk(response.Status()) {
		errorJson, err := json.Marshal(errorData)
		if err != nil {
			return err
		}
		return fmt.Errorf("response status for Get requests should be 200, got %d\nerror: %s", response.Status(), errorJson)
	} else {
		co.Set("objectId", responseData["objectId"])
	}
	co.clearChangedKeys()
	return nil
}

func (co *ClassObject) put() error {
	objectId, ok := co.Get("objectId")
	if ok == false {
		return fmt.Errorf("Object should have a `objectId` field to be puttable")
	}
	url := fmt.Sprintf("%s/classes/%s/%s", PARSE_URL, co.class, objectId)

	errorData := map[string]interface{}{}
	
	requestObj := co.putMap()
	responseData := map[string]interface{}{}
	response, err := session.Put(url, requestObj, &responseData, &errorData)
	if err != nil {
		return err
	}
	if statusOk(response.Status()) {
		co.SetFromMap(responseData)
	} else {
		errorJson, err := json.Marshal(errorData)
		if err != nil {
			return err
		}
		return fmt.Errorf("response status for Get requests should be 200, got %d\nerror: %s", response.Status(), errorJson)
	}
	co.clearChangedKeys()
	return nil
}

func (co *ClassObject) del() error {
	objectId, ok := co.Get("objectId")
	if ok == false {
		return fmt.Errorf("Object should have a `objectId` field to be deletable")
	}
	url := fmt.Sprintf("%s/classes/%s/%s", PARSE_URL, co.class, objectId)

	errorData := map[string]interface{}{}
	
	responseData := map[string]interface{}{}
	response, err := session.Delete(url, &responseData, &errorData)
	if err != nil {
		return err
	}
	if !statusOk(response.Status()) {
		errorJson, err := json.Marshal(errorData)
		if err != nil {
			return err
		}
		return fmt.Errorf("response status for Get requests should be 200, got %d\nerror: %s", response.Status(), errorJson)
	}
	return nil
}

/**
 *	Misc
 */

func statusOk(status int) bool {
	return status >= 200 && status < 300
}

func inArray(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
