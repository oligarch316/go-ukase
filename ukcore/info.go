package ukcore

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	paramsTagArgs = "ukargs"
	paramsTagFlag = "ukflag"
	paramsTagDoc  = "ukdoc"

	// TODO
	// paramsTagInline = "ukinline"
	// paramsTagValueSkip = "-"
)

var paramsInfoEmpty = ParamsInfo{
	TypeName: "EMPTY",
	Flags:    make(map[string]ParamsInfoFlag),
}

type ParamsInfo struct {
	TypeName string
	Args     *ParamsInfoArgs
	Flags    map[string]ParamsInfoFlag
}

type ParamsInfoArgs struct {
	FieldName  string
	FieldIndex []int
}

type ParamsInfoFlag struct {
	FieldName  string
	FieldIndex []int
	ParamsInfoFlagBool
}

type ParamsInfoFlagBool struct {
	IsBoolFlag    func() bool
	CheckBoolFlag func(string) bool
}

func ParamsInfoOf(v any) (ParamsInfo, error) {
	// TODO: Any need to descend indirects (pointer/interface) here?
	return NewParamsInfo(reflect.TypeOf(v))
}

func NewParamsInfo(typ reflect.Type) (ParamsInfo, error) {
	info := ParamsInfo{
		TypeName: typ.Name(),
		Flags:    make(map[string]ParamsInfoFlag),
	}

	if typ.Kind() != reflect.Struct {
		err := errorParams{TypeName: info.TypeName}
		return info, err.kind(typ.Kind())
	}

	for _, field := range reflect.VisibleFields(typ) {
		if err := info.loadField(field); err != nil {
			return info, err
		}
	}

	return info, nil
}

func (pi *ParamsInfo) loadField(field reflect.StructField) error {
	if _, ok := field.Tag.Lookup(paramsTagArgs); ok {
		return pi.loadArgsField(field)
	}

	return pi.loadFlagField(field)
}

func (pi *ParamsInfo) loadArgsField(field reflect.StructField) error {
	if pi.Args != nil {
		err := errorParams{TypeName: pi.TypeName}
		return err.argsConflict(pi.Args.FieldName, field.Name)
	}

	pi.Args = &ParamsInfoArgs{FieldName: field.Name, FieldIndex: field.Index}
	return nil
}

func (pi *ParamsInfo) loadFlagField(field reflect.StructField) error {
	flagInfo := ParamsInfoFlag{
		FieldName:          field.Name,
		FieldIndex:         field.Index,
		ParamsInfoFlagBool: pi.parseFlagBool(field),
	}

	// TODO:
	// "Denormalizing" multiple flag names here is convenient for decode
	// Probably pretty inconvenient when it comes to displaying help text
	// Cross that bridge...
	for _, flagName := range pi.parseFlagNames(field) {
		if extant, exists := pi.Flags[flagName]; exists {
			err := errorParams{TypeName: pi.TypeName}
			return err.flagConflict(extant.FieldName, field.Name)
		}

		pi.Flags[flagName] = flagInfo
	}

	return nil
}

func (ParamsInfo) parseFlagNames(field reflect.StructField) []string {
	// Prefer tag value(s) if available
	if tag, ok := field.Tag.Lookup(paramsTagFlag); ok {
		return strings.Fields(tag)
	}

	// Use field name with 1st rune lowercased otherwise
	r, size := utf8.DecodeRuneInString(field.Name)
	head := unicode.ToLower(r)
	tail := field.Name[size:]
	name := string(head) + tail
	return []string{name}
}

func (pi ParamsInfo) parseFlagBool(field reflect.StructField) ParamsInfoFlagBool {
	var boolInfo ParamsInfoFlagBool

	val := reflect.New(field.Type).Interface()

	type IsBool interface{ IsBoolFlag() bool }

	switch val := val.(type) {
	case IsBool:
		boolInfo.IsBoolFlag = val.IsBoolFlag
	case *bool:
		boolInfo.IsBoolFlag = pi.isBoolTrue
	default:
		boolInfo.IsBoolFlag = pi.isBoolFalse
	}

	type CheckBool interface{ CheckBoolFlag(string) bool }

	switch val := val.(type) {
	case CheckBool:
		boolInfo.CheckBoolFlag = val.CheckBoolFlag
	default:
		boolInfo.CheckBoolFlag = pi.checkBoolParse
	}

	return boolInfo
}

func (ParamsInfo) isBoolTrue() bool  { return true }
func (ParamsInfo) isBoolFalse() bool { return false }

func (ParamsInfo) checkBoolParse(str string) bool {
	_, err := strconv.ParseBool(str)
	return err == nil
}
