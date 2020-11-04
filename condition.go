package golik

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func getValueOf(path string, obj interface{}) reflect.Value {
	upperFirst := func(s string) string {
		arr := []rune(s)
		arr[0] = unicode.ToUpper(arr[0])
		return string(arr)
	}

	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	if value.IsValid() && (value.Kind() == reflect.Ptr || value.Kind() == reflect.Struct) {
		attrs := strings.SplitN(path, ".", 2)
		if len(attrs) == 0 {
			return reflect.ValueOf(nil)
		}
		attr := upperFirst(attrs[0])
		fld := value.FieldByName(attr)

		if len(attrs) > 1 {
			return getValueOf(attrs[1], fld.Interface())
		}

		return fld
	}
	return reflect.ValueOf(nil)
}

type Operator string

const (
	EQ Operator = "EQ" // equal
	NE          = "NE" // not equal
	CO          = "CO" // contains
	SW          = "SW" // start with
	EW          = "EW" // end with
	PR          = "PR" // present
	GT          = "GT" // greater
	GE          = "GE" // greater or equal
	LT          = "LT" // less
	LE          = "LE" // less or equal
)

var operators []Operator = []Operator{EQ, NE, CO, SW, EW, PR, GT, GE, LT, LE}

type Logical string

const (
	AND Logical = "AND"
	OR          = "OR"
)

type Condition interface {
	AND(right Condition) Logic
	OR(right Condition) Logic
	Command() string
	Check(obj interface{}) bool
}

type Operand interface {
	Condition
	Attribute() string
	Operator() Operator
	Value() interface{}
}

type Logic interface {
	Condition
	Left() Condition
	Logical() Logical
	Right() Condition
}

type LogicNot interface {
	Condition
	InnerNot() Condition
}

type Grouping interface {
	Condition
	InnerGroup() Condition
}

func AttributeCondition(name string, op Operator, value interface{}) Operand {
	/*
		if op == PR {
			return &prCondition{name}, nil
		}
	*/
	// TODO slice and map
	switch value.(type) {
	case string:
		return &stringCondition{
			attribute: name,
			operator:  op,
			value:     value.(string),
		}
	case bool:
		return &boolCondition{
			attribute: name,
			operator:  op,
			value:     value.(bool),
		}
	case int:
		return &intCondition{
			attribute: name,
			operator:  op,
			value:     value.(int),
		}
	case float64:
		return &float64Condition{
			attribute: name,
			operator:  op,
			value:     value.(float64),
		}
	default:
		//return nil, fmt.Errorf("Unsupportet datatype %T", value)
		return &prCondition{name}
	}
}

func IsEqual(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, EQ, value)
}

func IsNotEqual(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, NE, value)
}

func Contains(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, CO, value)
}

func StartWith(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, SW, value)
}

func EndsWith(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, EW, value)
}

func IsPresent(attribute string) Condition {
	return AttributeCondition(attribute, PR, nil)
}

func IsGreaterThan(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, GT, value)
}

func IsGreaterOrEqual(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, GE, value)
}

func IsLessThan(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, LT, value)
}

func IsLessOrEqual(attribute string, value interface{}) Condition {
	return AttributeCondition(attribute, LE, value)
}

func And(left Condition, right Condition) Logic {
	return &compoundCondition{
		left:    left,
		logical: AND,
		right:   right,
	}
}

func Or(left Condition, right Condition) Logic {
	return &compoundCondition{
		left:    left,
		logical: OR,
		right:   right,
	}
}

func Group(condition Condition) Grouping {
	return &groupCondition{condition}
}

func Not(condition Condition) LogicNot {
	return &notCondition{condition}
}

func EmptyCondition() Condition { return &emptyCondition{} }

type emptyCondition struct{}

func (ec *emptyCondition) AND(right Condition) Logic  { return And(ec, right) }
func (ec *emptyCondition) OR(right Condition) Logic   { return Or(ec, right) }
func (ec *emptyCondition) Command() string            { return "" }
func (ec *emptyCondition) Check(obj interface{}) bool { return true }

type prCondition struct {
	attribute string
}

func (c *prCondition) Command() string {
	return fmt.Sprintf("%v %v", c.attribute, PR)
}

func (c *prCondition) Check(obj interface{}) bool {
	value := getValueOf(c.attribute, obj)
	return value.IsValid() && !value.IsNil()
}

func (c *prCondition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *prCondition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *prCondition) Attribute() string {
	return c.attribute
}

func (c *prCondition) Operator() Operator {
	return PR
}

func (c *prCondition) Value() interface{} {
	return nil
}

type stringCondition struct {
	attribute string
	operator  Operator
	value     string
}

func (c *stringCondition) Command() string {
	return fmt.Sprintf("%v %v \"%v\"", c.attribute, c.operator, c.value)
}

func (c *stringCondition) Check(obj interface{}) bool {
	value := getValueOf(c.attribute, obj)
	if value.IsValid() && value.Kind() == reflect.String {
		v := value.String()
		switch c.operator {
		case EQ:
			return v == c.value
		case NE:
			return v != c.value
		case CO:
			return strings.Contains(v, c.value)
		case SW:
			return strings.HasPrefix(v, c.value)
		case EW:
			return strings.HasSuffix(v, c.value)
		case PR:
			return true
		case GT:
			return v > c.value
		case GE:
			return v >= c.value
		case LT:
			return v < c.value
		case LE:
			return v <= c.value
		}
	}
	return false
}

func (c *stringCondition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *stringCondition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *stringCondition) Attribute() string {
	return c.attribute
}

func (c *stringCondition) Operator() Operator {
	return c.operator
}

func (c *stringCondition) Value() interface{} {
	return c.value
}

type boolCondition struct {
	attribute string
	operator  Operator
	value     bool
}

func (c *boolCondition) Command() string {
	return fmt.Sprintf("%v %v %v", c.attribute, c.operator, c.value)
}

func (c *boolCondition) Check(obj interface{}) bool {
	value := getValueOf(c.attribute, obj)
	if value.IsValid() && value.Kind() == reflect.Bool {
		v := value.Bool()
		switch c.operator {
		case EQ:
			return c.value == v
		case NE:
			return c.value != v
		case CO:
			return c.value == v
		case SW:
			return c.value == v
		case EW:
			return c.value == v
		case PR:
			return true
		case GT:
			return c.value != v
		case GE:
			return true
		case LT:
			return c.value != v
		case LE:
			return true
		}
	}
	return false
}

func (c *boolCondition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *boolCondition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *boolCondition) Attribute() string {
	return c.attribute
}

func (c *boolCondition) Operator() Operator {
	return c.operator
}

func (c *boolCondition) Value() interface{} {
	return c.value
}

type intCondition struct {
	attribute string
	operator  Operator
	value     int
}

func (c *intCondition) Command() string {
	return fmt.Sprintf("%v %v %v", c.attribute, c.operator, c.value)
}

func (c *intCondition) Check(obj interface{}) bool {
	value := getValueOf(c.attribute, obj)
	if value.IsValid() && (value.Kind() == reflect.Int ||
		value.Kind() == reflect.Int8 ||
		value.Kind() == reflect.Int16 ||
		value.Kind() == reflect.Int32 ||
		value.Kind() == reflect.Int64 ||
		value.Kind() == reflect.Uint ||
		value.Kind() == reflect.Uint8 ||
		value.Kind() == reflect.Uint16 ||
		value.Kind() == reflect.Uint32 ||
		value.Kind() == reflect.Uint64) {
		cv := int64(c.value)
		v := value.Int()
		switch c.operator {
		case EQ:
			return v == cv
		case NE:
			return v != cv
		case CO:
			return strings.Contains(strconv.Itoa(int(v)), strconv.Itoa(c.value))
		case SW:
			return strings.HasPrefix(strconv.Itoa(int(v)), strconv.Itoa(c.value))
		case EW:
			return strings.HasSuffix(strconv.Itoa(int(v)), strconv.Itoa(c.value))
		case PR:
			return true
		case GT:
			return v > cv
		case GE:
			return v >= cv
		case LT:
			return v < cv
		case LE:
			return v <= cv
		}
	}
	return false
}

func (c *intCondition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *intCondition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *intCondition) Attribute() string {
	return c.attribute
}

func (c *intCondition) Operator() Operator {
	return c.operator
}

func (c *intCondition) Value() interface{} {
	return c.value
}

type float64Condition struct {
	attribute string
	operator  Operator
	value     float64
}

func (c *float64Condition) Command() string {
	return fmt.Sprintf("%v %v %v", c.attribute, c.operator, c.value)
}

func (c *float64Condition) Check(obj interface{}) bool {
	value := getValueOf(c.attribute, obj)
	if value.IsValid() && (value.Kind() == reflect.Float32 ||
		value.Kind() == reflect.Float64) {
		cv := c.value
		v := value.Float()
		switch c.operator {
		case EQ:
			return v == cv
		case NE:
			return v != cv
		case CO:
			return strings.Contains(fmt.Sprintf("%f", v), fmt.Sprintf("%f", cv))
		case SW:
			return strings.HasPrefix(fmt.Sprintf("%f", v), fmt.Sprintf("%f", cv))
		case EW:
			return strings.HasSuffix(fmt.Sprintf("%f", v), fmt.Sprintf("%f", cv))
		case PR:
			return true
		case GT:
			return v > cv
		case GE:
			return v >= cv
		case LT:
			return v < cv
		case LE:
			return v >= cv
		}
	}
	return false
}

func (c *float64Condition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *float64Condition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *float64Condition) Attribute() string {
	return c.attribute
}

func (c *float64Condition) Operator() Operator {
	return c.operator
}

func (c *float64Condition) Value() interface{} {
	return c.value
}

type compoundCondition struct {
	left    Condition
	logical Logical
	right   Condition
}

func (c *compoundCondition) Command() string {
	return fmt.Sprintf("%v %v %v", c.left.Command(), c.logical, c.right.Command())
}

func (c *compoundCondition) Check(obj interface{}) bool {
	switch c.logical {
	case AND:
		return c.left.Check(obj) && c.right.Check(obj)
	case OR:
		return c.left.Check(obj) || c.right.Check(obj)
	}
	return false
}

func (c *compoundCondition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *compoundCondition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *compoundCondition) Left() Condition {
	return c.left
}

func (c *compoundCondition) Logical() Logical {
	return c.logical
}

func (c *compoundCondition) Right() Condition {
	return c.right
}

type groupCondition struct {
	condition Condition
}

func (c *groupCondition) Command() string {
	return fmt.Sprintf("(%v)", c.condition.Command())
}

func (c *groupCondition) Check(obj interface{}) bool {
	return c.condition.Check(obj)
}

func (c *groupCondition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *groupCondition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *groupCondition) InnerGroup() Condition {
	return c.condition
}

type notCondition struct {
	condition Condition
}

func (c *notCondition) Command() string {
	return fmt.Sprintf("NOT %v", c.condition.Command())
}

func (c *notCondition) Check(obj interface{}) bool {
	return !c.condition.Check(obj)
}

func (c *notCondition) AND(right Condition) Logic {
	return And(c, right)
}

func (c *notCondition) OR(right Condition) Logic {
	return Or(c, right)
}

func (c *notCondition) InnerNot() Condition {
	return c.condition
}
