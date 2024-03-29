package ukcore

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	paramsTagArgs   = "ukargs"   // Arguments key
	paramsTagFlag   = "ukflag"   // Flag key
	paramsTagInline = "ukinline" // Inline struct key

	paramsTagValueSep  = " - " // Separator between (flag) names and doc
	paramsTagValueSkip = "-"   // Indicates (flag) field should be ignored
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

func ParamsInfoOf(v any) (ParamsInfo, error) {
	// TODO: Any need to descend indirects (pointer/interface) here?
	return NewParamsInfo(reflect.TypeOf(v))
}

func NewParamsInfo(typ reflect.Type) (ParamsInfo, error) {
	typeName, typeKind := typ.Name(), typ.Kind()
	if typeKind != reflect.Struct {
		err := errorParams{TypeName: typeName}
		return ParamsInfo{}, err.kind(typeKind)
	}

	flags := make(map[string]ParamsInfoFlag)
	info := ParamsInfo{TypeName: typeName, Flags: flags}
	return info, info.loadStruct(typ)
}

func (pi *ParamsInfo) loadStruct(typ reflect.Type) error {
	for _, field := range reflect.VisibleFields(typ) {
		if err := pi.loadField(field); err != nil {
			return err
		}
	}

	return nil
}

func (pi *ParamsInfo) loadField(field reflect.StructField) error {
	if tag, ok := field.Tag.Lookup(paramsTagInline); ok {
		return pi.loadTagInline(tag, field)
	}

	if tag, ok := field.Tag.Lookup(paramsTagArgs); ok {
		return pi.loadTagArgs(tag, field)
	}

	if tag, ok := field.Tag.Lookup(paramsTagFlag); ok {
		return pi.loadTagFlag(tag, field)
	}

	// Untagged field ⇒ treat as a flag
	flagName, flagDoc := pi.convertFlagName(field.Name), ""

	flagInfo := ParamsInfoFlag{Doc: flagDoc}
	flagInfo.loadField(field)

	return pi.addFlag(flagName, flagInfo)
}

func (pi *ParamsInfo) loadTagInline(tag string, field reflect.StructField) error {
	return errors.New("[TODO loadTagInline] not yet implemented")
}

func (pi *ParamsInfo) loadTagArgs(tag string, field reflect.StructField) error {
	if pi.Args != nil {
		err := errorParams{TypeName: pi.TypeName}
		return err.argsConflict(pi.Args.FieldName, field.Name)
	}

	pi.Args = &ParamsInfoArgs{Doc: strings.TrimSpace(tag)}
	pi.Args.loadField(field)
	return nil
}

func (pi *ParamsInfo) loadTagFlag(tag string, field reflect.StructField) error {
	head, tail, _ := strings.Cut(tag, paramsTagValueSep)
	head, tail = strings.TrimSpace(head), strings.TrimSpace(tail)

	flagNames, flagDoc := strings.Fields(head), tail

	flagInfo := ParamsInfoFlag{Doc: flagDoc}
	flagInfo.loadField(field)

	// TODO:
	// "Denormalizing" multiple flag names here is convenient for decode
	// Probably pretty inconvenient when it comes to displaying help text
	// Cross that bridge...
	for _, flagName := range flagNames {
		if err := pi.addFlag(flagName, flagInfo); err != nil {
			return err
		}
	}

	return nil
}

func (pi *ParamsInfo) addFlag(flagName string, flagInfo ParamsInfoFlag) error {
	if extant, exists := pi.Flags[flagName]; exists {
		err := errorParams{TypeName: pi.TypeName}
		return err.flagConflict(extant.FieldName, flagInfo.FieldName)
	}

	pi.Flags[flagName] = flagInfo
	return nil
}

func (ParamsInfo) convertFlagName(raw string) string {
	r, size := utf8.DecodeRuneInString(raw)
	head, tail := unicode.ToLower(r), raw[size:]
	return string(head) + tail
}

type ParamsInfoArgs struct {
	Doc        string
	FieldName  string
	FieldIndex []int
}

func (pia *ParamsInfoArgs) loadField(field reflect.StructField) {
	pia.FieldName = field.Name
	pia.FieldIndex = field.Index
}

type ParamsInfoFlag struct {
	Doc        string
	FieldName  string
	FieldIndex []int

	IsBoolFlag    func() bool
	CheckBoolFlag func(string) bool
}

func (pif *ParamsInfoFlag) loadField(field reflect.StructField) {
	pif.FieldName = field.Name
	pif.FieldIndex = field.Index

	val := reflect.New(field.Type).Interface()
	pif.loadBoolFuncs(val)
}

func (pif *ParamsInfoFlag) loadBoolFuncs(val any) {
	type IsBool interface{ IsBoolFlag() bool }
	type CheckBool interface{ CheckBoolFlag(string) bool }

	switch val := val.(type) {
	case IsBool:
		pif.IsBoolFlag = val.IsBoolFlag
	case *bool:
		pif.IsBoolFlag = pif.isBoolTrue
	default:
		pif.IsBoolFlag = pif.isBoolFalse
	}

	switch val := val.(type) {
	case CheckBool:
		pif.CheckBoolFlag = val.CheckBoolFlag
	default:
		pif.CheckBoolFlag = pif.checkBoolParse
	}
}

func (ParamsInfoFlag) isBoolTrue() bool  { return true }
func (ParamsInfoFlag) isBoolFalse() bool { return false }

func (ParamsInfoFlag) checkBoolParse(str string) bool {
	_, err := strconv.ParseBool(str)
	return err == nil
}
