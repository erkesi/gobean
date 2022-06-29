package ginjects

import (
	"fmt"
	"github.com/erkesi/gobean/glogs"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	graphtesta "github.com/erkesi/gobean/ginjects/graphtesta"
	graphtestb "github.com/erkesi/gobean/ginjects/graphtestb"
	"github.com/facebookgo/ensure"
)

func init() {
	// we rely on math.Rand in graph.objects() and this gives it some randomness.
	rand.Seed(time.Now().UnixNano())
}

type Answerable interface {
	Answer() int
}

type TypeAnswerStruct struct {
	answer  int
	private int
}

func (t *TypeAnswerStruct) Answer() int {
	return t.answer
}

type TypeNestedStruct struct {
	A *TypeAnswerStruct `inject:""`
}

func (t *TypeNestedStruct) Answer() int {
	return t.A.Answer()
}

// populate is a short-hand for populating a graph with the given incomplete
// object values.
func populate(values ...interface{}) error {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	for _, v := range values {
		if err := g.provide(&object{value: v}); err != nil {
			return err
		}
	}
	return g.populate()
}

func TestRequireTag(t *testing.T) {
	var v struct {
		A *TypeAnswerStruct
		B *TypeNestedStruct `inject:""`
	}

	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A != nil {
		t.Fatal("v.A is not nil")
	}
	if v.B == nil {
		t.Fatal("v.B is nil")
	}
}

type TypeWithNonPointerInject struct {
	A int `inject:""`
}

func TestErrorOnNonPointerInject(t *testing.T) {
	var a TypeWithNonPointerInject
	err := populate(&a)
	if err == nil {
		t.Fatalf("expected error for %+v", a)
	}

	const msg = "found inject tag on unsupported field A in type *ginjects.TypeWithNonPointerInject"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithNonPointerStructInject struct {
	A *int `inject:""`
}

func TestErrorOnNonPointerStructInject(t *testing.T) {
	var a TypeWithNonPointerStructInject
	err := populate(&a)
	if err == nil {
		t.Fatalf("expected error for %+v", a)
	}

	const msg = "found inject tag on unsupported field A in type *ginjects.TypeWithNonPointerStructInject"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestInjectSimple(t *testing.T) {
	var v struct {
		A *TypeAnswerStruct `inject:""`
		B *TypeNestedStruct `inject:""`
	}

	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A == nil {
		t.Fatal("v.A is nil")
	}
	if v.B == nil {
		t.Fatal("v.B is nil")
	}
	if v.B.A == nil {
		t.Fatal("v.B.A is nil")
	}
	if v.A != v.B.A {
		t.Fatal("got different instances of A")
	}
}

func TestDoesNotOverwrite(t *testing.T) {
	a := &TypeAnswerStruct{}
	var v struct {
		A *TypeAnswerStruct `inject:""`
		B *TypeNestedStruct `inject:""`
	}
	v.A = a
	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A != a {
		t.Fatal("original A was lost")
	}
	if v.B == nil {
		t.Fatal("v.B is nil")
	}
}

func TestPrivate(t *testing.T) {
	var v struct {
		A *TypeAnswerStruct `inject:"private"`
		B *TypeNestedStruct `inject:""`
	}

	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A == nil {
		t.Fatal("v.A is nil")
	}
	if v.B == nil {
		t.Fatal("v.B is nil")
	}
	if v.B.A == nil {
		t.Fatal("v.B.A is nil")
	}
	if v.A == v.B.A {
		t.Fatal("got the same A")
	}
}

type TypeWithJustColon struct {
	A *TypeAnswerStruct `inject:`
}

func TestTagWithJustColon(t *testing.T) {
	var a TypeWithJustColon
	err := populate(&a)
	if err == nil {
		t.Fatalf("expected error for %+v", a)
	}

	const msg = "unexpected tag format `inject:` for field A in type *ginjects.TypeWithJustColon"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithOpenQuote struct {
	A *TypeAnswerStruct `inject:"`
}

func TestTagWithOpenQuote(t *testing.T) {
	var a TypeWithOpenQuote
	err := populate(&a)
	if err == nil {
		t.Fatalf("expected error for %+v", a)
	}

	const msg = "unexpected tag format `inject:\"` for field A in type *ginjects.TypeWithOpenQuote"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestProvideWithFields(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	a := &TypeAnswerStruct{}
	err := g.provide(&object{value: a, fields: map[string]*object{}})
	ensure.NotNil(t, err)
	ensure.DeepEqual(t, err.Error(), "fields were specified on object \"*ginjects.TypeAnswerStruct\" when it was provided")
}

func TestProvideNonPointer(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	var i int
	err := g.provide(&object{value: i})
	if err == nil {
		t.Fatal("expected error")
	}

	const msg = "expected unnamed object value to be a pointer to a struct but got type int with value 0"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestProvideNonPointerStruct(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	var i *int
	err := g.provide(&object{value: i})
	if err == nil {
		t.Fatal("expected error")
	}

	const msg = "expected unnamed object value to be a pointer to a struct but got type *int with value <nil>"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestProvideTwoOfTheSame(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	a := TypeAnswerStruct{}
	err := g.provide(&object{value: &a})
	if err != nil {
		t.Fatal(err)
	}

	err = g.provide(&object{value: &a})
	if err == nil {
		t.Fatal("expected error")
	}

	const msg = "provided two unnamed instances of type *github.com/erkesi/gobean/ginjects.TypeAnswerStruct"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestProvideTwoOfTheSameWithpopulate(t *testing.T) {
	a := TypeAnswerStruct{}
	err := populate(&a, &a)
	if err == nil {
		t.Fatal("expected error")
	}

	const msg = "provided two unnamed instances of type *github.com/erkesi/gobean/ginjects.TypeAnswerStruct"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestProvideTwoWithTheSameName(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	const name = "foo"
	a := TypeAnswerStruct{}
	err := g.provide(&object{value: &a, name: name})
	if err != nil {
		t.Fatal(err)
	}

	err = g.provide(&object{value: &a, name: name})
	if err == nil {
		t.Fatal("expected error")
	}

	const msg = "provided two instances named foo"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestNamedInstanceWithDependencies(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	a := &TypeNestedStruct{}
	if err := g.provide(&object{value: a, name: "foo"}); err != nil {
		t.Fatal(err)
	}

	var c struct {
		A *TypeNestedStruct `inject:"foo"`
	}
	if err := g.provide(&object{value: &c}); err != nil {
		t.Fatal(err)
	}

	if err := g.populate(); err != nil {
		t.Fatal(err)
	}

	if c.A.A == nil {
		t.Fatal("c.A.A was not injected")
	}
}

func TestTwoNamedInstances(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	a := &TypeAnswerStruct{}
	b := &TypeAnswerStruct{}
	if err := g.provide(&object{value: a, name: "foo"}); err != nil {
		t.Fatal(err)
	}

	if err := g.provide(&object{value: b, name: "bar"}); err != nil {
		t.Fatal(err)
	}

	var c struct {
		A *TypeAnswerStruct `inject:"foo"`
		B *TypeAnswerStruct `inject:"bar"`
	}
	if err := g.provide(&object{value: &c}); err != nil {
		t.Fatal(err)
	}

	if err := g.populate(); err != nil {
		t.Fatal(err)
	}

	if c.A != a {
		t.Fatal("did not find expected c.A")
	}
	if c.B != b {
		t.Fatal("did not find expected c.B")
	}
}

type TypeWithMissingNamed struct {
	A *TypeAnswerStruct `inject:"foo"`
}

func TestTagWithMissingNamed(t *testing.T) {
	var a TypeWithMissingNamed
	err := populate(&a)
	if err == nil {
		t.Fatalf("expected error for %+v", a)
	}

	const msg = "did not find object named foo required by field A in type *ginjects.TypeWithMissingNamed"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestCompleteProvides(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	var v struct {
		A *TypeAnswerStruct `inject:""`
	}
	if err := g.provide(&object{value: &v, complete: true}); err != nil {
		t.Fatal(err)
	}

	if err := g.populate(); err != nil {
		t.Fatal(err)
	}
	if v.A != nil {
		t.Fatal("v.A was not nil")
	}
}

func TestCompleteNamedProvides(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	var v struct {
		A *TypeAnswerStruct `inject:""`
	}
	if err := g.provide(&object{value: &v, complete: true, name: "foo"}); err != nil {
		t.Fatal(err)
	}

	if err := g.populate(); err != nil {
		t.Fatal(err)
	}
	if v.A != nil {
		t.Fatal("v.A was not nil")
	}
}

type TypeInjectInterfaceMissing struct {
	Answerable Answerable `inject:""`
}

func TestInjectInterfaceMissing(t *testing.T) {
	var v TypeInjectInterfaceMissing
	err := populate(&v)
	if err == nil {
		t.Fatal("did not find expected error")
	}

	const msg = "found no assignable value for field Answerable in type *ginjects.TypeInjectInterfaceMissing"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeInjectInterface struct {
	Answerable Answerable        `inject:""`
	A          *TypeAnswerStruct `inject:""`
}

func TestInjectInterface(t *testing.T) {
	var v TypeInjectInterface
	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.Answerable == nil || v.Answerable != v.A {
		t.Fatalf(
			"expected the same but got Answerable = %T %+v / A = %T %+v",
			v.Answerable,
			v.Answerable,
			v.A,
			v.A,
		)
	}
}

type TypeWithInvalidNamedType struct {
	A *TypeNestedStruct `inject:"foo"`
}

func TestInvalidNamedInstanceType(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	a := &TypeAnswerStruct{}
	if err := g.provide(&object{value: a, name: "foo"}); err != nil {
		t.Fatal(err)
	}

	var c TypeWithInvalidNamedType
	if err := g.provide(&object{value: &c}); err != nil {
		t.Fatal(err)
	}

	err := g.populate()
	if err == nil {
		t.Fatal("did not find expected error")
	}

	const msg = "object named foo of type *ginjects.TypeNestedStruct is not assignable to field A (*ginjects.TypeAnswerStruct) in type *ginjects.TypeWithInvalidNamedType"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithInjectOnPrivateField struct {
	a *TypeAnswerStruct `inject:""`
}

func TestInjectOnPrivateField(t *testing.T) {
	var a TypeWithInjectOnPrivateField
	err := populate(&a)
	if err == nil {
		t.Fatal("did not find expected error")
	}

	const msg = "inject requested on unexported field a in type *ginjects.TypeWithInjectOnPrivateField"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithInjectOnPrivateInterfaceField struct {
	a Answerable `inject:""`
}

func TestInjectOnPrivateInterfaceField(t *testing.T) {
	var a TypeWithInjectOnPrivateField
	err := populate(&a)
	if err == nil {
		t.Fatal("did not find expected error")
	}

	const msg = "inject requested on unexported field a in type *ginjects.TypeWithInjectOnPrivateField"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeInjectPrivateInterface struct {
	Answerable Answerable        `inject:"private"`
	B          *TypeNestedStruct `inject:""`
}

func TestInjectPrivateInterface(t *testing.T) {
	var v TypeInjectPrivateInterface
	err := populate(&v)
	if err == nil {
		t.Fatal("did not find expected error")
	}

	const msg = "found private inject tag on interface field Answerable in type *ginjects.TypeInjectPrivateInterface"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeInjectTwoSatisfyInterface struct {
	Answerable Answerable        `inject:""`
	A          *TypeAnswerStruct `inject:""`
	B          *TypeNestedStruct `inject:""`
}

func TestInjectTwoSatisfyInterface(t *testing.T) {
	var v TypeInjectTwoSatisfyInterface
	err := populate(&v)
	if err == nil {
		t.Fatal("did not find expected error")
	}

	const msg = "found two assignable values for field Answerable in type *ginjects.TypeInjectTwoSatisfyInterface. one type *ginjects.TypeAnswerStruct with value &{0 0} and another type *ginjects.TypeNestedStruct with value"
	if !strings.HasPrefix(err.Error(), msg) {
		t.Fatalf("expected prefix:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeInjectNamedTwoSatisfyInterface struct {
	Answerable Answerable        `inject:""`
	A          *TypeAnswerStruct `inject:""`
	B          *TypeNestedStruct `inject:""`
}

func TestInjectNamedTwoSatisfyInterface(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	var v TypeInjectNamedTwoSatisfyInterface
	if err := g.provide(&object{name: "foo", value: &v}); err != nil {
		t.Fatal(err)
	}

	err := g.populate()
	if err == nil {
		t.Fatal("was expecting error")
	}

	const msg = "found two assignable values for field Answerable in type *ginjects.TypeInjectNamedTwoSatisfyInterface. one type *ginjects.TypeAnswerStruct with value &{0 0} and another type *ginjects.TypeNestedStruct with value"
	if !strings.HasPrefix(err.Error(), msg) {
		t.Fatalf("expected prefix:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithInjectNamedOnPrivateInterfaceField struct {
	a Answerable `inject:""`
}

func TestInjectNamedOnPrivateInterfaceField(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	var v TypeWithInjectNamedOnPrivateInterfaceField
	if err := g.provide(&object{name: "foo", value: &v}); err != nil {
		t.Fatal(err)
	}

	err := g.populate()
	if err == nil {
		t.Fatal("was expecting error")
	}

	const msg = "inject requested on unexported field a in type *ginjects.TypeWithInjectNamedOnPrivateInterfaceField"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithNonPointerNamedInject struct {
	A int `inject:"foo"`
}

func TestErrorOnNonPointerNamedInject(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	if err := g.provide(&object{name: "foo", value: 42}); err != nil {
		t.Fatal(err)
	}

	var v TypeWithNonPointerNamedInject
	if err := g.provide(&object{value: &v}); err != nil {
		t.Fatal(err)
	}

	if err := g.populate(); err != nil {
		t.Fatal(err)
	}

	if v.A != 42 {
		t.Fatalf("expected v.A = 42 but got %d", v.A)
	}
}

func TestInjectInline(t *testing.T) {
	var v struct {
		Inline struct {
			A *TypeAnswerStruct `inject:""`
			B *TypeNestedStruct `inject:""`
		} `inject:"inline"`
	}

	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.Inline.A == nil {
		t.Fatal("v.Inline.A is nil")
	}
	if v.Inline.B == nil {
		t.Fatal("v.Inline.B is nil")
	}
	if v.Inline.B.A == nil {
		t.Fatal("v.Inline.B.A is nil")
	}
	if v.Inline.A != v.Inline.B.A {
		t.Fatal("got different instances of A")
	}
}

func TestInjectInlineOnPointer(t *testing.T) {
	var v struct {
		Inline *struct {
			A *TypeAnswerStruct `inject:""`
			B *TypeNestedStruct `inject:""`
		} `inject:""`
	}

	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.Inline.A == nil {
		t.Fatal("v.Inline.A is nil")
	}
	if v.Inline.B == nil {
		t.Fatal("v.Inline.B is nil")
	}
	if v.Inline.B.A == nil {
		t.Fatal("v.Inline.B.A is nil")
	}
	if v.Inline.A != v.Inline.B.A {
		t.Fatal("got different instances of A")
	}
}

func TestInjectInvalidInline(t *testing.T) {
	var v struct {
		A *TypeAnswerStruct `inject:"inline"`
	}

	err := populate(&v)
	if err == nil {
		t.Fatal("was expecting an error")
	}

	const msg = `inline requested on non inlined field A in type *struct { A *ginjects.TypeAnswerStruct "inject:\"inline\"" }`
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

func TestInjectInlineMissing(t *testing.T) {
	var v struct {
		Inline struct {
			B *TypeNestedStruct `inject:""`
		} `inject:""`
	}

	err := populate(&v)
	if err == nil {
		t.Fatal("was expecting an error")
	}

	const msg = `inline struct on field Inline in type *struct { Inline struct { B *ginjects.TypeNestedStruct "inject:\"\"" } "inject:\"\"" } requires an explicit "inline" tag`
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithInlineStructWithPrivate struct {
	Inline struct {
		A *TypeAnswerStruct `inject:""`
		B *TypeNestedStruct `inject:""`
	} `inject:"private"`
}

func TestInjectInlinePrivate(t *testing.T) {
	var v TypeWithInlineStructWithPrivate
	err := populate(&v)
	if err == nil {
		t.Fatal("was expecting an error")
	}

	const msg = "cannot use private inject on inline struct on field Inline in type *ginjects.TypeWithInlineStructWithPrivate"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeWithStructValue struct {
	Inline TypeNestedStruct `inject:"inline"`
}

func TestInjectWithStructValue(t *testing.T) {
	var v TypeWithStructValue
	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.Inline.A == nil {
		t.Fatal("v.Inline.A is nil")
	}
}

type TypeWithNonpointerStructValue struct {
	Inline TypeNestedStruct `inject:"inline"`
}

func TestInjectWithNonpointerStructValue(t *testing.T) {
	var v TypeWithNonpointerStructValue
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	if err := g.provide(&object{value: &v}); err != nil {
		t.Fatal(err)
	}
	if err := g.populate(); err != nil {
		t.Fatal(err)
	}
	if v.Inline.A == nil {
		t.Fatal("v.Inline.A is nil")
	}
	n := len(g.objects())
	if n != 3 {
		t.Fatalf("expected 3 object in graph, got %d", n)
	}

}

func TestPrivateIsFollowed(t *testing.T) {
	var v struct {
		A *TypeNestedStruct `inject:"private"`
	}

	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A.A == nil {
		t.Fatal("v.A.A is nil")
	}
}

func TestDoesNotOverwriteInterface(t *testing.T) {
	a := &TypeAnswerStruct{}
	var v struct {
		A Answerable        `inject:""`
		B *TypeNestedStruct `inject:""`
	}
	v.A = a
	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A != a {
		t.Fatal("original A was lost")
	}
	if v.B == nil {
		t.Fatal("v.B is nil")
	}
}

func TestInterfaceIncludingPrivate(t *testing.T) {
	var v struct {
		A Answerable        `inject:""`
		B *TypeNestedStruct `inject:"private"`
		C *TypeAnswerStruct `inject:""`
	}
	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A == nil {
		t.Fatal("v.A is nil")
	}
	if v.B == nil {
		t.Fatal("v.B is nil")
	}
	if v.C == nil {
		t.Fatal("v.C is nil")
	}
	if v.A != v.C {
		t.Fatal("v.A != v.C")
	}
	if v.A == v.B {
		t.Fatal("v.A == v.B")
	}
}

func TestInjectMap(t *testing.T) {
	var v struct {
		A map[string]int `inject:"private"`
	}
	if err := populate(&v); err != nil {
		t.Fatal(err)
	}
	if v.A == nil {
		t.Fatal("v.A is nil")
	}
}

type TypeInjectWithMapWithoutPrivate struct {
	A map[string]int `inject:""`
}

func TestInjectMapWithoutPrivate(t *testing.T) {
	var v TypeInjectWithMapWithoutPrivate
	err := populate(&v)
	if err == nil {
		t.Fatalf("expected error for %+v", v)
	}

	const msg = "inject on map field A in type *ginjects.TypeInjectWithMapWithoutPrivate must be named or private"
	if err.Error() != msg {
		t.Fatalf("expected:\n%s\nactual:\n%s", msg, err.Error())
	}
}

type TypeForObjectString struct {
	A *TypeNestedStruct `inject:"foo"`
	B *TypeNestedStruct `inject:""`
}

func TestObjectString(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	a := &TypeNestedStruct{}
	if err := g.provide(&object{value: a, name: "foo"}); err != nil {
		t.Fatal(err)
	}

	var c TypeForObjectString
	if err := g.provide(&object{value: &c}); err != nil {
		t.Fatal(err)
	}

	if err := g.populate(); err != nil {
		t.Fatal(err)
	}

	var actual []string
	for _, o := range g.objects() {
		actual = append(actual, fmt.Sprint(o))
	}

	ensure.SameElements(t, actual, []string{
		"\"*ginjects.TypeAnswerStruct\"",
		"\"*ginjects.TypeNestedStruct named foo\"",
		"\"*ginjects.TypeNestedStruct\"",
		"\"*ginjects.TypeForObjectString\"",
	})
}

type TypeForGraphObjects struct {
	TypeNestedStruct `inject:"inline"`
	A                *TypeNestedStruct `inject:"foo"`
	E                struct {
		B *TypeNestedStruct `inject:""`
	} `inject:"inline"`
}

func TestGraphObjects(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	err := g.provide(
		&object{value: &TypeNestedStruct{}, name: "foo"},
		&object{value: &TypeForGraphObjects{}},
	)
	ensure.Nil(t, err)
	ensure.Nil(t, g.populate())

	var actual []string
	for _, o := range g.objects() {
		actual = append(actual, fmt.Sprint(o))
	}

	ensure.SameElements(t, actual, []string{
		"\"*ginjects.TypeAnswerStruct\"",
		"\"*ginjects.TypeForGraphObjects\"",
		"\"*ginjects.TypeNestedStruct named foo\"",
		"\"*ginjects.TypeNestedStruct\"",
		`"*struct { B *ginjects.TypeNestedStruct "inject:\"\"" }"`,
	})
}

type logger struct {
	Expected []string
	T        testing.TB
	next     int
}

func (l *logger) Debugf(f string, v ...interface{}) {
	actual := fmt.Sprintf(f, v...)
	if l.next == len(l.Expected) {
		l.T.Fatalf(`unexpected log "%s"`, actual)
	}
	expected := l.Expected[l.next]
	if actual != expected {
		l.T.Fatalf(`expected log "%s" got "%s"`, expected, actual)
	}
	l.next++
}

type TypeForLoggingInterface interface {
	Foo()
}

type TypeForLoggingCreated struct{}

func (t TypeForLoggingCreated) Foo() {}

type TypeForLoggingEmbedded struct {
	TypeForLoggingCreated      *TypeForLoggingCreated  `inject:""`
	TypeForLoggingInterface    TypeForLoggingInterface `inject:""`
	TypeForLoggingCreatedNamed *TypeForLoggingCreated  `inject:"name_for_logging"`
	Map                        map[string]string       `inject:"private"`
}

type TypeForLogging struct {
	TypeForLoggingEmbedded `inject:"inline"`
	TypeForLoggingCreated  *TypeForLoggingCreated `inject:""`
}

func InjectLogging(t *testing.T) {
	glogs.Init(&logger{
		Expected: []string{
			"provided *ginjects.TypeForLoggingCreated named name_for_logging",
			"provided *ginjects.TypeForLogging",
			"provided embedded *ginjects.TypeForLoggingEmbedded",
			"created *ginjects.TypeForLoggingCreated",
			"assigned newly created *ginjects.TypeForLoggingCreated to field TypeForLoggingCreated in *ginjects.TypeForLogging",
			"assigned existing *ginjects.TypeForLoggingCreated to field TypeForLoggingCreated in *ginjects.TypeForLoggingEmbedded",
			"assigned *ginjects.TypeForLoggingCreated named name_for_logging to field TypeForLoggingCreatedNamed in *ginjects.TypeForLoggingEmbedded",
			"made map for field Map in *ginjects.TypeForLoggingEmbedded",
			"assigned existing *ginjects.TypeForLoggingCreated to interface field TypeForLoggingInterface in *ginjects.TypeForLoggingEmbedded",
		},
		T: t,
	})
	g := &graph{refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{}}
	var v TypeForLogging

	err := g.provide(
		&object{value: &TypeForLoggingCreated{}, name: "name_for_logging"},
		&object{value: &v},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.populate(); err != nil {
		t.Fatal(err)
	}
}

type TypeForNamedWithUnnamedDepSecond struct{}

type TypeForNamedWithUnnamedDepFirst struct {
	TypeForNamedWithUnnamedDepSecond *TypeForNamedWithUnnamedDepSecond `inject:""`
}

type TypeForNamedWithUnnamed struct {
	TypeForNamedWithUnnamedDepFirst *TypeForNamedWithUnnamedDepFirst `inject:""`
}

func TestForNamedWithUnnamed(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	var v TypeForNamedWithUnnamed

	err := g.provide(
		&object{value: &v, name: "foo"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.populate(); err != nil {
		t.Fatal(err)
	}
	if v.TypeForNamedWithUnnamedDepFirst == nil {
		t.Fatal("expected TypeForNamedWithUnnamedDepFirst to be populated")
	}
	if v.TypeForNamedWithUnnamedDepFirst.TypeForNamedWithUnnamedDepSecond == nil {
		t.Fatal("expected TypeForNamedWithUnnamedDepSecond to be populated")
	}
}

func TestForSameNameButDifferentPackage(t *testing.T) {
	g := &graph{
		refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{},
	}
	err := g.provide(
		&object{value: &graphtesta.Foo{}},
		&object{value: &graphtestb.Foo{}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.populate(); err != nil {
		t.Fatal(err)
	}
}

func TestStructTagExtract(t *testing.T) {
	cases := []struct {
		Name  string // Input name
		Tag   string // Input tag
		Found bool   // Expected found status
		Value string // Expected value
		Error bool   // Indicates if an error is expected
	}{
		{
			Name:  "inject",
			Tag:   `inject:`,
			Error: true,
		},
		{
			Name:  "inject",
			Tag:   `inject:"`,
			Error: true,
		},
		{
			Name:  "inject",
			Tag:   `inject:""`,
			Found: true,
		},
		{
			Name:  "inject",
			Tag:   `inject:"a"`,
			Found: true,
			Value: "a",
		},
		{
			Name:  "inject",
			Tag:   ` inject:"a"`,
			Found: true,
			Value: "a",
		},
		{
			Name: "inject",
			Tag:  `  `,
		},
		{
			Name:  "inject",
			Tag:   `inject:"\"a"`,
			Found: true,
			Value: `"a`,
		},
	}

	for _, e := range cases {
		found, value, err := structTagExtract(e.Name, e.Tag)

		if !e.Error && err != nil {
			t.Fatalf("unexpected error %s for case %+v", err, e)
		}
		if e.Error && err == nil {
			t.Fatalf("did not get expected error for case %+v", e)
		}

		if found != e.Found {
			if e.Found {
				t.Fatalf("did not find value when expecting to %+v", e)
			} else {
				t.Fatalf("found value when not expecting to %+v", e)
			}
		}

		if value != e.Value {
			t.Fatalf(`found unexpected value "%s" for %+v`, value, e)
		}
	}
}
