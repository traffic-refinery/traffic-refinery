package counters

import (
	"errors"
	"reflect"
)

var typeRegistry = make(map[string]reflect.Type)

func registerType(typedNil interface{}) {
	t := reflect.TypeOf(typedNil).Elem()
	typeRegistry[t.Name()] = t
}

// AvailableCounters is a structure containing available Counter
// types
type AvailableCounters struct {
	registryByName map[string]reflect.Type
	registryById   map[int]reflect.Type
	idToName       map[int]string
	nameToId       map[string]int
}

// isCounter checks whether type t implements a counter interface
func isCounter(t reflect.Type) bool {
	if _, ok := reflect.New(t).Elem().Addr().Interface().(Counter); !ok {
		return false
	} else {
		return true
	}
}

// Build iterates over the counter names that the program plans on using.
// If no type with such name is found or if the data type is not a counter
// interface an error is raised
func (ac *AvailableCounters) Build(counters []string) (map[string]int, error) {
	ac.registryByName = make(map[string]reflect.Type)
	ac.registryById = make(map[int]reflect.Type)
	ac.idToName = make(map[int]string)
	ac.nameToId = make(map[string]int)
	lastCode := 0
	for _, counter := range counters {
		if val, found := typeRegistry[counter]; found {
			if !isCounter(typeRegistry[counter]) {
				return nil, errors.New("counter " + counter + " is not of the correct type " + reflect.New(typeRegistry[counter]).Elem().Type().Name())
			}
			ac.idToName[lastCode] = counter
			ac.nameToId[counter] = lastCode
			ac.registryById[lastCode] = val
			ac.registryByName[counter] = val
		} else {
			return nil, errors.New("counter " + counter + " does not exist")
		}
		lastCode++
	}
	return ac.nameToId, nil
}

// InstantiateByName instantiates a new Counter of type with name counterName
func (ac *AvailableCounters) InstantiateByName(counterName string) (Counter, error) {
	v := reflect.New(ac.registryByName[counterName]).Elem().Addr().Interface()
	return v.(Counter), nil
}

// InstantiateById instantiates a new Counter of the type associated with code
// counterId
func (ac *AvailableCounters) InstantiateById(counterId int) (Counter, error) {
	v := reflect.New(ac.registryById[counterId]).Elem().Addr().Interface()
	return v.(Counter), nil
}
