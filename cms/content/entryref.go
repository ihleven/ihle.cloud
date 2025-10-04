package content

import (
	"encoding/json"
	"net/url"
	"strings"
)

type Resolv int

const (
	RSLV_REFS Resolv = 1 << iota
	RSLV_LINKS
	RSLV_ASSETS
	RSLV_CB
	RSLV_ALL  Resolv = RSLV_REFS | RSLV_LINKS | RSLV_ASSETS
	RSLV_NONE Resolv = 0
)

var _resolvValue = map[string]Resolv{
	"refs":   RSLV_REFS,
	"links":  RSLV_LINKS,
	"assets": RSLV_ASSETS,
	"cb":     RSLV_CB,
	"all":    RSLV_ALL,
}

func ParseResolv(inputvalues []string) Resolv {

	v := RSLV_NONE

	for _, value := range inputvalues {
		for _, split := range strings.Split(value, ",") {
			if resolv, ok := _resolvValue[split]; ok {
				v |= resolv
			}
		}
	}

	return v
}

// {
// 	  "referenceWithPathFieldName":
// 	    "refkey": "entries/usp/usp-columns/en-UK.json" ||  "pages:about-us/abta-protected"
// 	  }
// }
// {
// 	  "referenceWithIDFieldName":
// 	    "ref_id": "pages:about-us/abta-protected"
// 	  }
// }

// Ref ist ein Verweis auf einen anderen Eintrag
// Refkey ist entweder der Pfad oder die ID des Eintrags.
// Modes:
// 1. Default: anstelle der Ref wird der referenzierte Entry ausgeliefert. Es sind aber lediglich die Metadaten aus der Ref enthalten (kodiert im refkey).
//
//	{
//		 "referenceWithPathFieldName": {
//		   "refkey": "entries/usp/usp-columns/en-UK.json" ||  "pages:about-us/abta-protected"
//	      "content": {
//	       ...
//	      }
//		   }
//	  }
//	}
//
// 2. Entry: anstelle der Ref wird der referenzierte Entry ausgeliefert. Es sind aber lediglich die Metadaten aus der Ref enthalten.
//
//	{
//		"referenceWithPathFieldName":
//			"refkey": "entries/usp/usp-columns/en-UK.json" ||  "pages:about-us/abta-protected"
//	     "entry": {
//	       "path": ...
//	       "content": {
//	         ...
//	       }
//	     }
//		}
//	}
//
// 3. Content: anstelle der Ref wird der EntryContent ausgeliefert. So erzeugte Entries sind readlonly, da sie nicht mehr referenzierbar sind und dem Unmarshaler nicht mehr alle benötigten Information zur Verfügung stehen und der einen Fehler liefert.
//
//		"referenceWithPathFieldName": {
//		  "value": " ... "
//	   "nested": {
//		  }
//	}
type EntryRef struct {
	// Key        string      `json:"refkey,omitempty"`
	Path       string `json:"-"`
	Collection string `json:"-"`
	ID         string `json:"-"`
	Mode       string `json:"refmode,omitempty"`
	*Entry     `json:",omitempty"`
}

// Wenn entry in ref enthalten, dann rendern, ansonsten lediglich den key
func (r *EntryRef) MarshalJSON() ([]byte, error) {

	// im Content-Mode wird bewusst auf eine Rücküberführung verzichtet.
	if r.Mode == "Content" && r.Entry != nil {
		return json.Marshal(r.Entry.Content)
	}

	var refkey url.URL
	if r.Path != "" {
		refkey.Path = r.Path
	} else {
		refkey.Scheme = r.Collection
		refkey.Opaque = r.ID
	}

	// create struct to avoid infinite recursion
	aux := struct {
		Key  string `json:"refkey,omitempty"`
		Mode string `json:"refmode,omitempty"`
		// Path       string      `json:"path,omitempty"`
		// Collection string      `json:"collection,omitempty"`
		// ID         string      `json:"id,omitempty"`
		Type    string      `json:"type,omitempty"`
		Content interface{} `json:"content,omitempty"`
	}{Key: refkey.String(), Mode: r.Mode} //, Path: r.Path, Collection: r.Collection, ID: r.ID}

	if r.Entry != nil {
		aux.Type = r.Entry.Type
		aux.Content = r.Entry.Content
	}

	return json.Marshal(&aux)
}

// Zunächst auf einen enthaltenen refkey prüfen.
// wenn das nicht der Fall ist, einen Entry parsen.
func (r *EntryRef) UnmarshalJSON(data []byte) error {

	var aux struct {
		Key  string `json:"refkey,omitempty"`
		Mode string `json:"refmode,omitempty"`
	}

	err := json.Unmarshal(data, &aux)
	if err != nil {
		// fmt.Printf("Ref.UnmarshalJSON error => %s\n", err)
		return err
	}

	// if aux.Key != "" {
	url, err := url.Parse(aux.Key)
	if err != nil {
		return err
	}
	r.Path = url.Path
	r.Collection = url.Scheme
	r.ID = url.Opaque
	r.Mode = aux.Mode
	// } else {
	// 	r.Mode = "Content"
	// }

	var e Entry
	err = json.Unmarshal(data, &e)
	if err == nil {
		// dereferenzierter entry
		r.Entry = &e
	}

	return nil
}

// type EntryRef struct {
// 	Path   string `json:"refkey,omitempty"`
// 	*Entry `json:",omitempty"`
// }

// // Zunächst auf einen enthaltenen refkey prüfen.
// // wenn das nicht der Fall ist, einen Entry parsen.
// func (b *EntryRef) UnmarshalJSON(data []byte) error {
// 	var aux struct {
// 		Ref string `json:"refkey"`
// 	}
// 	err := json.Unmarshal(data, &aux)
// 	if err != nil {
// 		return err
// 	}
// 	if aux.Ref != "" {
// 		b.Path = aux.Ref
// 		return nil
// 	}

// 	var e Entry
// 	err = json.Unmarshal(data, &e)
// 	if err != nil {
// 		return err
// 	}
// 	b.Path = e.Path
// 	b.Entry = &e
// 	return nil
// }

// // Wenn entry in ref enthalten, dann rendern, ansonsten lediglich den key
// func (b *EntryRef) MarshalJSON() ([]byte, error) {

// 	if b.Entry != nil {
// 		return json.Marshal(b.Entry)
// 	}
// 	var aux struct {
// 		Ref string `json:"refkey"`
// 	}
// 	aux.Ref = b.Path
// 	return json.Marshal(&aux)
// }
