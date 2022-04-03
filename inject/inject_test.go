package inject

import (
	"fmt"
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
	fmt.Println("B.init" + b.name)
}

func (b *B) Close() {
	fmt.Println("B.close")
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

func Test_Inject(t *testing.T) {
	ProvideByValue(&A{}, ProvideWithPriority(99))
	ProvideByValue(&B{name: "hel"}, ProvideWithPriority(100))
	ProvideByName("b", &B{name: "hel2"}, ProvideWithPriority(101))
	c := &C{}
	ProvideByValue(c)
	Init()
	t.Log(PrintObjects())
	c = ObtainByType(c).(*C)
	t.Log(c.B.name)
	t.Log(c.B1.name)
}
