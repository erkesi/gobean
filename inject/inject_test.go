package inject

import (
	"fmt"
	"github.com/erkesi/gobean/log"
	"reflect"
	"testing"
)

type A struct {
}

func (a *A) Init() {
	fmt.Println("A.init")
}

func (a *A) Close() {
	fmt.Println("A.close")
}

type B struct {
	name string
}

func (b *B) Init() {
	fmt.Println("B.init " + b.name)
}

func (b *B) Close() {
	fmt.Println("B.close " + b.name)
}

func (b B) String() string {
	return b.name
}

type C struct {
	A  *A `inject:""`
	B  *B `inject:""`
	B1 *B `inject:"b"`
}

func (c *C) Init() {
	fmt.Println("C.init")
}

func (c *C) Close() {
	fmt.Println("C.close")
}

type Log struct {
}

func (l Log) Debugf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func Test_Inject(t *testing.T) {
	log.Init(Log{})
	ProvideByValue(&A{}, ProvideWithPriority(99))
	ProvideByValue(&B{name: "unName"}, ProvideWithPriority(100))
	ProvideByName("b", &B{name: "named"}, ProvideWithPriority(101))
	c := &C{}
	ProvideByValue(c)

	Init()

	if reflect.DeepEqual(PrintObjects(), []string{"*inject.B named b", "*inject.B", "*inject.A", "*inject.C"}) {
		t.Fatal("Init() objects order error")
		return
	}

	if c.B.name != "unName" {
		t.Fatal("err")
		return
	}

	if c.B1.name != "named" {
		t.Fatal("err")
		return
	}

	b := ObtainByName("b").(*B)

	if b.name != "named" {
		t.Fatal("err")
		return
	}

	Close()
}
