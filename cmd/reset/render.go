package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"sort"
)

type renderer struct {
	pkg     *types.Package
	imports map[string]*types.Package
	aliases map[string]string
	used    map[string]bool
	body    bytes.Buffer
	dynamic string
}

func renderPackage(pkg *types.Package, targets []resetTarget) ([]byte, error) {
	r := &renderer{
		pkg: pkg, imports: make(map[string]*types.Package),
		aliases: make(map[string]string), used: make(map[string]bool),
	}
	for _, name := range pkg.Scope().Names() {
		r.used[name] = true
	}
	for _, target := range targets {
		for i := 0; i < target.named.TypeParams().Len(); i++ {
			r.used[target.named.TypeParams().At(i).Obj().Name()] = true
		}
	}
	receiver := r.identifier("rs")
	for _, target := range targets {
		var params []string
		for i := 0; i < target.named.TypeParams().Len(); i++ {
			params = append(params, target.named.TypeParams().At(i).Obj().Name())
		}
		fmt.Fprintf(&r.body, "\n// Reset restores %s to its initial state.\n", target.name)
		fmt.Fprintf(&r.body, "func (%s *%s%s) Reset() {\nif %s == nil { return }\n", receiver, target.name, typeArguments(params), receiver)
		for i := 0; i < target.fields.NumFields(); i++ {
			field := target.fields.Field(i)
			if field.Name() == "_" {
				continue
			}
			if err := r.field(receiver+"."+field.Name(), field.Type()); err != nil {
				return nil, fmt.Errorf("%s.%s: %w", target.name, field.Name(), err)
			}
		}
		r.body.WriteString("}\n")
	}
	if r.dynamic != "" {
		r.dynamicHelper()
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "%s\n\npackage %s\n", generatedMark, pkg.Name())
	if len(r.imports) != 0 {
		out.WriteString("\nimport (\n")
		var paths []string
		for path := range r.imports {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		for _, path := range paths {
			alias := r.aliases[path]
			if alias != r.imports[path].Name() {
				fmt.Fprintf(&out, "%s ", alias)
			}
			fmt.Fprintf(&out, "%q\n", path)
		}
		out.WriteString(")\n")
	}
	out.Write(r.body.Bytes())
	data, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}
	return data, nil
}

func (r *renderer) field(expr string, fieldType types.Type) error {
	return r.fieldValue(expr, fieldType, make(map[types.Type]bool))
}

func (r *renderer) fieldValue(expr string, fieldType types.Type, active map[types.Type]bool) error {
	if active[fieldType] {
		r.dynamicValue(expr)
		return nil
	}
	active[fieldType] = true
	defer delete(active, fieldType)
	if _, ok := types.Unalias(fieldType).(*types.TypeParam); ok {
		r.dynamicValue(expr)
		return nil
	}
	switch field := fieldType.Underlying().(type) {
	case *types.Basic:
		switch {
		case field.Info()&types.IsBoolean != 0:
			fmt.Fprintf(&r.body, "%s = false\n", expr)
		case field.Info()&types.IsString != 0:
			fmt.Fprintf(&r.body, "%s = \"\"\n", expr)
		case field.Info()&types.IsNumeric != 0:
			fmt.Fprintf(&r.body, "%s = 0\n", expr)
		case field.Kind() == types.UnsafePointer:
			fmt.Fprintf(&r.body, "%s = nil\n", expr)
		default:
			return fmt.Errorf("unsupported basic type %s", fieldType)
		}
	case *types.Slice:
		fmt.Fprintf(&r.body, "%s = %s[:0]\n", expr, operand(expr))
	case *types.Map:
		fmt.Fprintf(&r.body, "clear(%s)\n", expr)
	case *types.Pointer:
		fmt.Fprintf(&r.body, "if %s != nil {\n", expr)
		if err := r.fieldValue("*"+expr, field.Elem(), active); err != nil {
			return err
		}
		r.body.WriteString("}\n")
	case *types.Struct:
		method, _, _ := types.LookupFieldOrMethod(fieldType, true, r.pkg, "Reset")
		if method, ok := method.(*types.Func); ok && isReset(method) {
			fmt.Fprintf(&r.body, "%s.Reset()\n", operand(expr))
		} else {
			r.zero(expr, fieldType)
		}
	case *types.Array:
		index := r.identifier("resetIndex")
		fmt.Fprintf(&r.body, "for %s := range %s {\n", index, expr)
		if err := r.fieldValue(operand(expr)+"["+index+"]", field.Elem(), active); err != nil {
			return err
		}
		r.body.WriteString("}\n")
	case *types.Interface, *types.Chan, *types.Signature:
		fmt.Fprintf(&r.body, "%s = nil\n", expr)
	default:
		return fmt.Errorf("unsupported type %s", fieldType)
	}
	return nil
}

func operand(expr string) string {
	if expr[0] == '*' {
		return "(" + expr + ")"
	}
	return expr
}

func (r *renderer) dynamicValue(expr string) {
	if r.dynamic == "" {
		r.dynamic = r.identifier("resetValue")
	}
	reflect := r.qualifier(types.NewPackage("reflect", "reflect"))
	fmt.Fprintf(&r.body, "%s(%s.ValueOf(&(%s)).Elem())\n", r.dynamic, reflect, expr)
}

// Type parameters and recursive pointer types need runtime type information.
// Ordinary fields are reset by direct assignments and method calls.
func (r *renderer) dynamicHelper() {
	reflect := r.qualifier(types.NewPackage("reflect", "reflect"))
	fmt.Fprintf(&r.body, `
func %[1]s(value %[2]s.Value) {
	switch value.Kind() {
	case %[2]s.Pointer:
		if !value.IsNil() {
			%[1]s(value.Elem())
		}
	case %[2]s.Slice:
		value.SetLen(0)
	case %[2]s.Map:
		value.Clear()
	case %[2]s.Array:
		for i := 0; i < value.Len(); i++ {
			%[1]s(value.Index(i))
		}
	case %[2]s.Struct:
		if value.CanAddr() {
			if resetter, ok := value.Addr().Interface().(interface{ Reset() }); ok {
				resetter.Reset()
				return
			}
		}
		value.SetZero()
	default:
		value.SetZero()
	}
}
`, r.dynamic, reflect)
}

func (r *renderer) zero(expr string, fieldType types.Type) {
	name := r.identifier("zero")
	fmt.Fprintf(&r.body, "{\nvar %s %s\n%s = %s\n}\n", name, types.TypeString(fieldType, r.qualifier), expr, name)
}

func (r *renderer) qualifier(pkg *types.Package) string {
	if pkg.Path() == r.pkg.Path() {
		return ""
	}
	if alias, ok := r.aliases[pkg.Path()]; ok {
		return alias
	}
	alias := r.identifier(pkg.Name())
	r.imports[pkg.Path()] = pkg
	r.aliases[pkg.Path()] = alias
	return alias
}

func (r *renderer) identifier(base string) string {
	name := base
	for i := 1; r.used[name]; i++ {
		name = fmt.Sprintf("%s%d", base, i)
	}
	r.used[name] = true
	return name
}
