package ginjects

import (
	"context"
	"errors"
	"fmt"
	"github.com/erkesi/gobean/glogs"
	"github.com/facebookgo/ensure"
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

type Log struct {
}

func (l Log) Debugf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func (l Log) Errorf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
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
