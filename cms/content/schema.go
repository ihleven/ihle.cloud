package content

import (
	"encoding/json"
	"fmt"
	"log"

	dynamicstruct "github.com/ompluscator/dynamic-struct"

	"github.com/swaggest/jsonschema-go"
)

// In a future version of the CMS we want to support generating content type via something like json schema
// for being able to generate types on the fly without a redeployment
// the feature war prioritized for a later release

type tODO struct {
	Amount float64  `json:"amount" minimum:"10.5" example:"20.6" required:"true"`
	Abc    string   `json:"abc" pattern:"[abc]"`
	_      struct{} `additionalProperties:"false"`                   // Tags of unnamed field are applied to parent schema.
	_      struct{} `title:"My Struct" description:"Holds my data."` // Multiple unnamed fields can be used.
}

func jsonSchema(t interface{}) json.RawMessage {

	reflector := jsonschema.Reflector{}

	schema, err := reflector.Reflect(t)
	if err != nil {
		log.Fatal(err)
	}

	j, err := json.MarshalIndent(schema, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(j))

	return json.RawMessage(j)
}

func dynstrct() {
	types := map[string]interface{}{
		"int":    0,
		"string": "",
	}

	instance := dynamicstruct.NewStruct().
		AddField("Integer", types["int"], `json:"int"`).
		AddField("Text", types["string"], `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		Build().
		New()

	data := []byte(`
		{
			"int": 123,
			"someText": "example",
			"double": 123.45,
			"Boolean": true,
			"Slice": [1, 2, 3],
			"Anonymous": "avoid to read"
		}
		`)

	err := json.Unmarshal(data, &instance)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))

}
