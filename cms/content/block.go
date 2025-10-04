package content

import (
	"encoding/json"
	"reflect"
)

type Blok struct {
	ID      int    `json:"id"`
	UUID    string `json:"uuid,omitempty"`
	Content Bloker `json:"content"`
}

type Bloker interface {
	Clone() Bloker
}

func (b *Blok) Clone() Blok {

	blok := Blok{
		ID:      b.ID,
		UUID:    b.UUID,
		Content: b.Content.Clone(),
	}
	// fmt.Printf("Blok.Clone() -> %p === %p Content: %p === %p\n", &b, &blok, b.Content, blok.Content)

	return blok
}

func (b *Blok) MarshalJSON() ([]byte, error) {

	return json.MarshalIndent(&struct {
		ID int `json:"id"`
		// UUID      string          `json:"uuid,omitempty"`
		Component string      `json:"component"`
		Content   interface{} `json:"content"`
		// Schema    json.RawMessage `json:"schema,omitempty"`
	}{
		ID:        b.ID,
		Component: componentName(b.Content),
		Content:   b.Content,
		// Schema:    jsonSchema(b.Content),
	}, "", "    ")
}

func (b *Blok) UnmarshalJSON(data []byte) error {

	var aux struct {
		ID        int             `json:"id"`
		Component string          `json:"component"`
		Content   json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	b.ID = aux.ID

	if t, err := InstantiateComponent(aux.Component); err == nil {
		if err := json.Unmarshal(aux.Content, t); err != nil {
			return err
		}
		b.Content = t
	}

	return nil
}

// Component structs have this type embedded
type Component struct{}

// helper for marshaling components
func componentName(b interface{}) string {
	t := reflect.TypeOf(b)
	if t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t != nil && t.Kind() == reflect.Struct {
		return t.Field(0).Tag.Get("name")
	}
	return "invalid"
}
