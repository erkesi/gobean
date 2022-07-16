// Package inject provides a reflect based injector. A large application built
// with dependency injection in mind will typically involve the boring work of
// setting up the object graph. This library attempts to take care of this
// boring work by creating and connecting the various objects. Its use involves
// you seeding the object graph with some (possibly incomplete) objects, where
// the underlying types have been tagged for injection. Given this, the
// library will populate the objects creating new ones as necessary. It uses
// singletons by default, supports optional private instances as well as named
// instances.
//
// It works using Go's reflection package and is inherently limited in what it
// can do as opposed to a code-gen system with respect to private fields.
//
// The usage pattern for the library involves struct tags. It requires the tag
// format used by the various standard libraries, like json, xml etc. It
// involves tags in one of the three forms below:
//
//     `inject:""`
//     `inject:"private"`
//     `inject:"dev logger"`
//
// The first no value syntax is for the common case of a singleton dependency
// of the associated type. The second triggers creation of a private instance
// for the associated type. Finally the last form is asking for a named
// dependency called "dev logger".
package ginjects

import (
	"errors"
	"fmt"
	"github.com/erkesi/gobean/glogs"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// An object in the graph.
type object struct {
	index        int
	priority     int
	value        interface{}
	name         string             // Optional
	complete     bool               // If true, the value will be considered complete
	fields       map[string]*object // Populated with the field names that were injected and their corresponding *object.
	reflectType  reflect.Type
	reflectValue reflect.Value
	private      bool // If true, the value will not be used and will only be populated
	created      bool // If true, the object was created by us
	embedded     bool // If true, the object is an embedded struct provided internally
	initFn       func()
	closeFn      func()
}

// String representation suitable for human consumption.
func (o *object) String() string {
	var builder strings.Builder
	builder.WriteString(`"`)
	builder.WriteString(o.reflectType.String())
	if o.name != "" {
		builder.WriteString(fmt.Sprintf(" named %s", o.name))
	}
	builder.WriteString(`"`)
	return builder.String()
}

func (o *object) addDep(g *graph, fieldName string, dep *object) {
	if o.fields == nil {
		o.fields = make(map[string]*object)
	}
	o.fields[fieldName] = dep
	g.edges = append(g.edges, Edge{{
		index:    dep.index,
		priority: dep.priority,
	}, {
		index:    o.index,
		priority: o.priority,
	}})
}

// The graph of objects.
type graph struct {
	sync.Once
	index                     int
	unnamed                   []*object
	unnamedType               map[reflect.Type]interface{}
	named                     map[string]*object
	index2Object              map[int]*object
	edges                     []Edge
	nodes                     []EdgeNode
	allNodes                  []EdgeNode
	refType2InjectFieldIndies map[reflect.Type]map[int]struct{}
}

// provide objects to the graph. The object documentation describes
// the impact of various fields.
func (g *graph) provide(objects ...*object) error {
	for _, o := range objects {
		g.index += 1
		o.index = g.index
		g.allNodes = append(g.allNodes, EdgeNode{
			index:    o.index,
			priority: o.priority,
		})
		if g.index2Object == nil {
			g.index2Object = make(map[int]*object)
		}
		g.index2Object[o.index] = o
		if lc, ok := o.value.(ObjectInit); ok {
			o.initFn = lc.Init
		}
		if lc, ok := o.value.(ObjectClose); ok {
			o.closeFn = lc.Close
		}
		o.reflectType = reflect.TypeOf(o.value)
		o.reflectValue = reflect.ValueOf(o.value)

		if o.fields != nil {
			return fmt.Errorf(
				"ginjects: fields were specified on object %s when it was provided",
				o,
			)
		}

		if o.name == "" {
			if !isStructPtr(o.reflectType) {
				return fmt.Errorf(
					"ginjects: expected unnamed object value to be a pointer to a struct but got type %s "+
						"with value %v",
					o.reflectType,
					o.value,
				)
			}

			if !o.private {
				if g.unnamedType == nil {
					g.unnamedType = make(map[reflect.Type]interface{})
				}

				if _, ok := g.unnamedType[o.reflectType]; ok {
					return fmt.Errorf(
						"ginjects: provided two unnamed instances of type *%s.%s",
						o.reflectType.Elem().PkgPath(), o.reflectType.Elem().Name(),
					)
				}
				g.unnamedType[o.reflectType] = o.value
			}
			g.unnamed = append(g.unnamed, o)
		} else {
			if g.named == nil {
				g.named = make(map[string]*object)
			}

			if g.named[o.name] != nil {
				return fmt.Errorf("ginjects: provided two instances named %s", o.name)
			}
			g.named[o.name] = o
		}

		if glogs.Log != nil {
			if o.created {
				glogs.Log.Debugf("ginjects: created %s", o)
			} else if o.embedded {
				glogs.Log.Debugf("ginjects: provided embedded %s", o)
			} else {
				glogs.Log.Debugf("ginjects: provided %s", o)
			}
		}
	}
	return nil
}

// populate the incomplete objects.
func (g *graph) populate() error {
	for _, o := range g.named {
		if o.complete {
			continue
		}

		if err := g.populateExplicit(o); err != nil {
			return err
		}
	}

	// We append and modify our slice as we go along, so we don't use a standard
	// range loop, and do a single pass thru each object in our graph.
	i := 0
	for {
		if i == len(g.unnamed) {
			break
		}

		o := g.unnamed[i]
		i++

		if o.complete {
			continue
		}

		if err := g.populateExplicit(o); err != nil {
			return err
		}
	}

	// A Second pass handles injecting Interface values to ensure we have created
	// all concrete types first.
	for _, o := range g.unnamed {
		if o.complete {
			continue
		}

		if err := g.populateUnnamedInterface(o); err != nil {
			return err
		}
	}

	for _, o := range g.named {
		if o.complete {
			continue
		}

		if err := g.populateUnnamedInterface(o); err != nil {
			return err
		}
	}
	nodes, circles, err := Toposort(g.edges, g.allNodes)
	if err != nil {
		nodes := make([]string, 0, len(circles))
		for _, node := range circles {
			nodes = append(nodes, g.index2Object[node.index].String())
		}
		if len(nodes) > 0 {
			return fmt.Errorf("ginjects: object depend inject graph is circle, %s", strings.Join(nodes, " >> "))
		}
		return err
	}
	g.nodes = nodes
	return nil
}

func (g *graph) populateExplicit(o *object) error {
	// Ignore named value types.
	if o.name != "" && !isStructPtr(o.reflectType) {
		return nil
	}

StructLoop:
	for i := 0; i < o.reflectValue.Elem().NumField(); i++ {
		field := o.reflectValue.Elem().Field(i)
		fieldType := field.Type()
		fieldTag := o.reflectType.Elem().Field(i).Tag
		fieldName := o.reflectType.Elem().Field(i).Name
		tag, err := parseTag(string(fieldTag))
		if err != nil {
			return fmt.Errorf(
				"ginjects: unexpected tag format `%s` for field %s in type %s",
				string(fieldTag),
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// Skip fields without a tag.
		if tag == nil {
			continue
		}
		g.addInjectFieldIndex(o.reflectType, i)
		// Cannot be used with unexported fields.
		if !field.CanSet() {
			return fmt.Errorf(
				"ginjects: inject requested on unexported field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// Inline tag on anything besides a struct is considered invalid.
		if tag.Inline && fieldType.Kind() != reflect.Struct {
			return fmt.Errorf(
				"ginjects: inline requested on non inlined field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// Don't overwrite existing values.
		if !isNilOrZero(field, fieldType) {
			continue
		}

		// Named injects must have been explicitly provided.
		if tag.Name != "" {
			existing := g.named[tag.Name]
			if existing == nil {
				return fmt.Errorf(
					"ginjects: did not find object named %s required by field %s in type %s",
					tag.Name,
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			if !existing.reflectType.AssignableTo(fieldType) {
				return fmt.Errorf(
					"ginjects: object named %s of type %s is not assignable to field %s (%s) in type %s",
					tag.Name,
					fieldType,
					o.reflectType.Elem().Field(i).Name,
					existing.reflectType,
					o.reflectType,
				)
			}

			field.Set(reflect.ValueOf(existing.value))
			if glogs.Log != nil {
				glogs.Log.Debugf(
					"ginjects: assigned %s to field %s in %s",
					existing,
					o.reflectType.Elem().Field(i).Name,
					o,
				)
			}
			o.addDep(g, fieldName, existing)
			continue StructLoop
		}

		// Inline struct values indicate we want to traverse into it, but not
		// inject itself. We require an explicit "inline" tag for this to work.
		if fieldType.Kind() == reflect.Struct {
			if tag.Private {
				return fmt.Errorf(
					"ginjects: cannot use private inject on inline struct on field %s in type %s",
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			if !tag.Inline {
				return fmt.Errorf(
					"ginjects: inline struct on field %s in type %s requires an explicit \"inline\" tag",
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			err := g.provide(&object{
				value:    field.Addr().Interface(),
				private:  true,
				embedded: o.reflectType.Elem().Field(i).Anonymous,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Interface injection is handled in a second pass.
		if fieldType.Kind() == reflect.Interface {
			continue
		}

		// Maps are created and required to be private.
		if fieldType.Kind() == reflect.Map {
			if !tag.Private {
				return fmt.Errorf(
					"ginjects: inject on map field %s in type %s must be named or private",
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			field.Set(reflect.MakeMap(fieldType))
			if glogs.Log != nil {
				glogs.Log.Debugf(
					"ginjects: made map for field %s in %s",
					o.reflectType.Elem().Field(i).Name,
					o,
				)
			}
			continue
		}

		// Can only inject Pointers from here on.
		if !isStructPtr(fieldType) {
			return fmt.Errorf(
				"ginjects: found inject tag on unsupported field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// Unless it's a private inject, we'll look for an existing instance of the
		// same type.
		if !tag.Private {
			for _, existing := range g.unnamed {
				if existing.private {
					continue
				}
				if existing.reflectType.AssignableTo(fieldType) {
					field.Set(reflect.ValueOf(existing.value))
					if glogs.Log != nil {
						glogs.Log.Debugf(
							"ginjects: assigned existing %s to field %s in %s",
							existing,
							o.reflectType.Elem().Field(i).Name,
							o,
						)
					}
					o.addDep(g, fieldName, existing)
					continue StructLoop
				}
			}
		}

		newValue := reflect.New(fieldType.Elem())
		newObject := &object{
			value:   newValue.Interface(),
			private: tag.Private,
			created: true,
		}

		// Add the newly ceated object to the known set of objects.
		err = g.provide(newObject)
		if err != nil {
			return err
		}

		// Finally assign the newly created object to our field.
		field.Set(newValue)
		if glogs.Log != nil {
			glogs.Log.Debugf(
				"ginjects: assigned newly created %s to field %s in %s",
				newObject,
				o.reflectType.Elem().Field(i).Name,
				o,
			)
		}
		o.addDep(g, fieldName, newObject)
	}
	return nil
}

func (g *graph) populateUnnamedInterface(o *object) error {
	// Ignore named value types.
	if o.name != "" && !isStructPtr(o.reflectType) {
		return nil
	}

	for i := 0; i < o.reflectValue.Elem().NumField(); i++ {
		field := o.reflectValue.Elem().Field(i)
		fieldType := field.Type()
		fieldTag := o.reflectType.Elem().Field(i).Tag
		fieldName := o.reflectType.Elem().Field(i).Name
		tag, err := parseTag(string(fieldTag))
		if err != nil {
			return fmt.Errorf(
				"ginjects: unexpected tag format `%s` for field %s in type %s",
				string(fieldTag),
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// Skip fields without a tag.
		if tag == nil {
			continue
		}
		g.addInjectFieldIndex(o.reflectType, i)
		// We only handle interface injection here. Other cases including errors
		// are handled in the first pass when we inject pointers.
		if fieldType.Kind() != reflect.Interface {
			continue
		}

		// Interface injection can't be private because we can't instantiate new
		// instances of an interface.
		if tag.Private {
			return fmt.Errorf(
				"ginjects: found private inject tag on interface field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// Don't overwrite existing values.
		if !isNilOrZero(field, fieldType) {
			continue
		}

		// Named injects must have already been handled in populateExplicit.
		if tag.Name != "" {
			panic(fmt.Sprintf("ginjects: unhandled named instance with name %s", tag.Name))
		}

		// Find one, and only one assignable value for the field.
		var found *object
		for _, existing := range g.unnamed {
			if existing.private {
				continue
			}
			if existing.reflectType.AssignableTo(fieldType) {
				if found != nil {
					return fmt.Errorf(
						"ginjects: found two assignable values for field %s in type %s. one type "+
							"%s with value %v and another type %s with value %v",
						o.reflectType.Elem().Field(i).Name,
						o.reflectType,
						found.reflectType,
						found.value,
						existing.reflectType,
						existing.reflectValue,
					)
				}
				found = existing
				field.Set(reflect.ValueOf(existing.value))
				if glogs.Log != nil {
					glogs.Log.Debugf(
						"ginjects: assigned existing %s to interface field %s in %s",
						existing,
						o.reflectType.Elem().Field(i).Name,
						o,
					)
				}
				o.addDep(g, fieldName, existing)
			}
		}

		// If we didn't find an assignable value, we're missing something.
		if found == nil {
			return fmt.Errorf(
				"ginjects: found no assignable value for field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}
	}
	return nil
}

// objects returns all known objects, named as well as unnamed. The returned
// elements are not in a stable order.
func (g *graph) objects() []*object {
	objects := make([]*object, 0, len(g.unnamed)+len(g.named))
	for _, node := range g.nodes {
		o := g.index2Object[node.index]
		if !o.embedded {
			objects = append(objects, o)
		}
	}
	return objects
}

var (
	injectOnly    = &tag{}
	injectPrivate = &tag{Private: true}
	injectInline  = &tag{Inline: true}
)

type tag struct {
	Name    string
	Inline  bool
	Private bool
}

func parseTag(t string) (*tag, error) {
	found, value, err := structTagExtract("inject", t)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	if value == "" {
		return injectOnly, nil
	}
	if value == "inline" {
		return injectInline, nil
	}
	if value == "private" {
		return injectPrivate, nil
	}
	return &tag{Name: value}, nil
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func isNilOrZero(v reflect.Value, t reflect.Type) bool {
	switch v.Kind() {
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(t).Interface())
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
}

func (g *graph) addInjectFieldIndex(refType reflect.Type, i int) {
	m, ok := g.refType2InjectFieldIndies[refType]
	if !ok {
		m = map[int]struct{}{}
		g.refType2InjectFieldIndies[refType] = m
	}
	m[i] = struct{}{}
}

var errInvalidTag = errors.New("ginjects: invalid tag")

// structTagExtract the quoted value for the given name returning it if it is found. The
// found boolean helps differentiate between the "empty and found" vs "empty
// and not found" nature of default empty strings.
func structTagExtract(name, tag string) (found bool, value string, err error) {
	for tag != "" {
		// skip leading space
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// scan to colon.
		// a space or a quote is a syntax error
		i = 0
		for i < len(tag) && tag[i] != ' ' && tag[i] != ':' && tag[i] != '"' {
			i++
		}
		if i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			return false, "", errInvalidTag
		}
		foundName := string(tag[:i])
		tag = tag[i+1:]

		// scan quoted string to find value
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return false, "", errInvalidTag
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if foundName == name {
			value, err := strconv.Unquote(qvalue)
			if err != nil {
				return false, "", err
			}
			return true, value, nil
		}
	}
	return false, "", nil
}
