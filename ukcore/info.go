package ukcore

import (
	"errors"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"
)

const (
	tagKeyArgs = "ukargs"
	tagKeyFlag = "ukflag"
	tagKeyDoc  = "ukdoc"
)

var EmptyParamsInfo = ParamsInfo{Flags: make(map[string]FlagInfo)}

func defaultIsBool() bool { return false }

func defaultCheckBool(str string) bool {
	_, err := strconv.ParseBool(str)
	return err == nil
}

type ArgsInfo struct{ Index []int }

type FlagInfo struct {
	Index []int

	IsBoolFlag    func() bool
	CheckBoolFlag func(string) bool
}

type ParamsInfo struct {
	Args  *ArgsInfo
	Flags map[string]FlagInfo
}

func ParamsInfoOf(v any) (ParamsInfo, error) {
	return NewParamsInfo(reflect.TypeOf(v))
}

func NewParamsInfo(typ reflect.Type) (ParamsInfo, error) {
	info := ParamsInfo{Flags: make(map[string]FlagInfo)}

	if typ.Kind() != reflect.Struct {
		return info, errors.New("[TODO NewParamsInfo] typ was not of kind struct")
	}

	for _, field := range reflect.VisibleFields(typ) {
		if err := info.loadField(field); err != nil {
			return info, err
		}
	}

	return info, nil
}

func (pi *ParamsInfo) loadField(field reflect.StructField) error {
	if _, ok := field.Tag.Lookup(tagKeyArgs); ok {
		return pi.loadArgsField(field)
	}

	return pi.loadFlagField(field)
}

func (pi *ParamsInfo) loadArgsField(field reflect.StructField) error {
	if pi.Args != nil {
		return errors.New("[TODO loadFieldArgs] detected multiple args fields")
	}

	pi.Args = &ArgsInfo{Index: field.Index}
	return nil
}

func (pi *ParamsInfo) loadFlagField(field reflect.StructField) error {
	type IsBool interface{ IsBoolFlag() bool }
	type CheckBool interface{ CheckBoolFlag(string) bool }

	flagName := pi.parseFlagName(field)
	if _, exists := pi.Flags[flagName]; exists {
		return errors.New("[TODO loadFlagField] detected conflicting flag fields")
	}

	flagInfo := FlagInfo{
		Index:         field.Index,
		IsBoolFlag:    defaultIsBool,
		CheckBoolFlag: defaultCheckBool,
	}

	// TODO: Do I need to worry about `field.Type` being an indirect type (Pointer/Interface)?
	tempVal := reflect.New(field.Type).Interface()

	switch tempT := tempVal.(type) {
	case IsBool:
		flagInfo.IsBoolFlag = tempT.IsBoolFlag
	case *bool:
		flagInfo.IsBoolFlag = func() bool { return true }
	}

	if checkBool, ok := tempVal.(CheckBool); ok {
		flagInfo.CheckBoolFlag = checkBool.CheckBoolFlag
	}

	pi.Flags[flagName] = flagInfo
	return nil
}

func (ParamsInfo) parseFlagName(field reflect.StructField) string {
	// Prefer tag value if available
	if tag, ok := field.Tag.Lookup(tagKeyFlag); ok {
		return tag
	}

	// Use field name with 1st rune lowercased otherwise
	r, size := utf8.DecodeRuneInString(field.Name)
	head, tail := string(unicode.ToLower(r)), field.Name[size:]
	return head + tail
}
