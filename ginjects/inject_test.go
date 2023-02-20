package ginjects

import (
	"context"
	"errors"
	"fmt"
	"github.com/erkesi/gobean/glogs"
	"github.com/facebookgo/ensure"
	"reflect"
	"testing"
)

type A struct {
	Name string
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
	A1    *A    `inject:"private"`
	A2    *A    `inject:"private"`
	B     *B    `inject:""`
	B1    *B    `inject:"b"`
	Error error `inject:""`
}

func (c *C) Init() {
	fmt.Println("C.init")
}

func (c *C) Close() {
	fmt.Println("C.close")
}

func Test_1(t *testing.T) {
	// Golang program to illustrate
	// reflect.New() Function
	type Geek struct {
		A int `tag1:"First Tag" tag2:"Second Tag"`
		B string
	}

	greeting := "GeeksforGeeks"
	f := Geek{A: 10, B: "Number"}

	gVal := reflect.ValueOf(greeting)

	fmt.Println(gVal.Interface())

	gpVal := reflect.ValueOf(&greeting)
	gpVal.Elem().SetString("Articles")
	fmt.Println(greeting)

	fType := reflect.TypeOf(f)
	fVal := reflect.New(fType)
	fmt.Printf("%p\n", fVal.Interface())
	fVal.Elem().Field(0).SetInt(20)
	fVal.Elem().Field(1).SetString("Number")
	f2 := fVal.Elem().Interface().(Geek)
	fmt.Printf("%+v, %d, %s\n", f2, f2.A, f2.B)
	fVal = reflect.New(fType)
	fmt.Printf("%p\n", fVal.Interface())
}

func Test_Inject(t *testing.T) {
	glogs.Init(Log{})

	ProvideByValue(&A{}, ProvideWithPriority(99))
	ProvideByValue(&B{name: "unName"}, ProvideWithPriority(100))
	ProvideByName("b", &B{name: "named"}, ProvideWithPriority(101))
	errObj := errors.New("err")
	ProvideByValue(errObj)

	c := &C{}
	ProvideByValue(c)

	Init()

	ensure.SameElements(t, PrintObjects(), []string{"\"*ginjects.B named b\"", "\"*ginjects.B\"", "\"*ginjects.A\"", "\"*errors.errorString\"", "\"*ginjects.A\"", "\"*ginjects.A\"", "\"*ginjects.C\""})

	// obtain by struct type
	c1 := ObtainByType(&C{}).(*C)
	if c != c1 {
		t.Fatal("err")
		return
	}
	if c1.B.name != "unName" {
		t.Fatal("err")
		return
	}

	if c1.B1.name != "named" {
		t.Fatal("err")
		return
	}
	if c1.A1 == c1.A2 {
		t.Fatalf("c1.A1(%p) == c1.A2(%p)", c1.A1, c1.A2)
		return
	}
	// obtain by interface type
	var err error
	err = ObtainByType(&err).(error)
	if err != errObj {
		t.Fatal("err")
		return
	}

	// obtain by name
	b := ObtainByName("b").(*B)

	if b.name != "named" {
		t.Fatal("err")
		return
	}

	Close()
}

type Log struct {
}

func (l Log) Debugf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func (l Log) Errorf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}
