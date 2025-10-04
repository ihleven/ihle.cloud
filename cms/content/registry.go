package content

import (
	"fmt"
	"log"
	"reflect"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

var typeRegistry = map[string]reflect.Type{}

var componentRegistry = map[string]reflect.Type{}

func Register(b interface{}) {

	t := reflect.TypeOf(b)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return
	}

	if t.Kind() != reflect.Struct {
		return
	}

	ptrtype := reflect.New(t).Type()

	if field, ok := t.FieldByName("ContentType"); ok {
		if name := field.Tag.Get("type"); name != "" {

			if _, ok := componentRegistry[name]; ok {
				log.Fatalf("Error in Register content type: %q: already exists", name)
			}

			clonerReflectType := reflect.TypeOf((*cloner)(nil)).Elem()

			if !ptrtype.Implements(clonerReflectType) {
				log.Fatalf("Error in Register component: %q: does not implement cloner interface", name)
			}

			typeRegistry[name] = reflect.TypeOf(b)
		}
	}

	if field, ok := t.FieldByName("Component"); ok {
		if name := field.Tag.Get("name"); name != "" {

			if _, ok := componentRegistry[name]; ok {
				log.Fatalf("Error in Register component: %q: already exists", name)
			}

			BlokerReflectType := reflect.TypeOf((*Bloker)(nil)).Elem()

			if !ptrtype.Implements(BlokerReflectType) {
				log.Fatalf("Error in Register component: %q: does not implement Bloker interface", name)
			}

			componentRegistry[name] = reflect.TypeOf(b)
		}
	}
}

// Instantiate returns a Pointer to the component or type identified by `name` as interface.
func Instantiate(name string) (interface{}, error) {
	// fmt.Printf("Instantiate %s %v\n", name, typeRegistry[name])
	if name == "" {
		return nil, fmt.Errorf("empty type name")
	}

	t, ok := componentRegistry[name]
	if !ok {
		t, ok = typeRegistry[name]
		if !ok {
			return nil, fmt.Errorf("unrecognized type name: %s", name)
		}
	}
	// fmt.Println("registry,", typeRegistry)
	return reflect.New(t).Interface(), nil
}

func InstantiateComponent(name string) (Bloker, error) {
	// fmt.Printf("Instantiate %s\n", name)
	t, ok := componentRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unrecognized type name: %s", name)
	}
	iface := reflect.New(t).Interface()
	if bloker, ok := iface.(Bloker); ok {
		return bloker, nil
	}
	return nil, errors.New("Error in InstantiateComponent: %s doesn't bloker interface", name)
}
