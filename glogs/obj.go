package glogs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type Obj struct {
	data   interface{}
	encode bool
}

func NewObj(v interface{}, encodes ...bool) *Obj {
	encode := false
	if len(encodes) > 0 {
		encode = encodes[0]
	}
	return &Obj{data: v, encode: encode}
}

func (f Obj) String() string {
	bs, _ := json.Marshal(f.data)
	s := string(bs)
	if f.encode {
		s = strconv.QuoteToASCII(s)
	}
	ret := fmt.Sprintf("Obj: %s(%s)", reflect.Indirect(reflect.ValueOf(f.data)).Type().Name(), s)
	return ret
}
