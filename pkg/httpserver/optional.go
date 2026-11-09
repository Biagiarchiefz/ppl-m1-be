package httpserver

import (
	"bytes"
	"encoding/json"
)

// Optional membedakan field JSON yang tidak dikirim (Set=false) dengan field
// bernilai null (Set=true, Value=nil). Dipakai untuk body PATCH.
type Optional[T any] struct {
	Set   bool
	Value *T
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(data, []byte("null")) {
		o.Value = nil
		return nil
	}
	return json.Unmarshal(data, &o.Value)
}
