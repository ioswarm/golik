package filter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	attrRegex  = regexp.MustCompile(`(?P<attribute>.*)\s+(?P<operator>[a-zA-Z]{2})\s+(?P<value>.*)`) // TODO no whitespaces in attribute
	valueRegex = regexp.MustCompile(`(?P<float>[+-]?[0-9]+\.[0-9]+)|(?P<int>[+-]?[0-9]+)|(?P<bool>true|false)|(?:['"](?P<string>.*)['"])`)
)

func parseValue(s string) (interface{}, error) {
	vmatch := valueRegex.FindStringSubmatch(s)
	for i, name := range valueRegex.SubexpNames() {
		vm := vmatch[i]
		if i != 0 && name != "" && vm != "" {
			switch name {
			case "bool":
				return strconv.ParseBool(vm)
			case "float":
				return strconv.ParseFloat(vm, 64)
			case "int":
				return strconv.Atoi(vm)
			case "string":
				return vm, nil
			}
		}
	}
	return nil, fmt.Errorf("Could not parse value %v", s)
}

func checkOpertor(s string) (Operator, error) {
	for _, op := range operators {
		if op == Operator(strings.ToUpper(s)) {
			return op, nil
		}
	}
	return "", fmt.Errorf("Operator %v is unknown", s)
}

func parseAttribute(s string) (Condition, error) {
	if attrRegex.MatchString(s) {
		match := attrRegex.FindStringSubmatch(s)
		vals := make(map[string]string)
		for i, name := range attrRegex.SubexpNames() {
			if i != 0 && name != "" {
				vals[name] = match[i]
			}
		}

		op, oerr := checkOpertor(vals["operator"])
		if oerr != nil {
			return nil, oerr
		}

		ival, iverr := parseValue(vals["value"])
		if iverr != nil {
			return nil, iverr
		}

		return AttributeCondition(vals["attribute"], op, ival), nil
	}
	return nil, fmt.Errorf("Condition %s is not valid", s)
}

func parseGroup(s string) (Condition, error) {
	indicator := 0
	idx := 0
	for i, r := range s {
		if r == '(' {
			indicator++
		}
		if r == ')' {
			indicator--
		}
		if indicator == 0 {
			idx = i
			break
		}
	}

	cond, err := Parse(s[1:idx])
	if err != nil {
		return nil, err
	}
	var result Condition = Group(cond)
	
	if idx+1 < len(s) {
		subcond, suberr := Parse(strings.TrimSpace(s[idx+1:]))
		if suberr != nil {
			return nil, suberr
		}

		switch subcond.(type) {
		case *compoundCondition:
			cc := subcond.(*compoundCondition)
			cc.left = result
			result = cc
		default:
			return nil, fmt.Errorf("Missing logical operator AND/OR at position %v", idx+1)
		}
	}
	return result, nil
}

func parseLogicalRight(s string) (Condition, error) {
	lower := strings.ToLower(s)

	cond, err := Parse(s[(strings.Index(s, " ")+1):])
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(lower, "or") {
		return Or(nil, cond), nil
	}
	return And(nil, cond), nil
}

func parseLogical(s string) (Condition, error) {
	lower := strings.ToLower(s)
	idx := strings.Index(lower, "and")
	oridx := strings.Index(lower, "or")
	op := AND
	if idx < 0 && oridx < 0 {
		return nil, fmt.Errorf("Need AND/OR operator in %v", s)
	}
	if idx < 0 || (oridx >= 0 && idx > oridx) {
		idx = oridx
		op = OR
	}
	if idx == 0 {
		return parseLogicalRight(s)
	}
	left, lerr := Parse(s[:idx-1]) 
	if lerr != nil {
		return nil, lerr
	}
	right, rerr := Parse(s[idx+len(op)+1:])
	if rerr != nil {
		return nil, rerr
	}
	return &compoundCondition{
		left: left,
		logical: op,
		right: right,
	}, nil
}

func parseNot(s string) (Condition, error) {
	cond, err := Parse(strings.TrimSpace(s[3:]))
	if err != nil {
		return nil, err
	}
	return Not(cond), nil
}

func Parse(s string) (Condition, error) {
	lower := strings.ToLower(strings.TrimSpace(s))
	if (strings.HasPrefix(lower, "not")) {
		return parseNot(strings.TrimSpace(s))
	}
	if (strings.HasPrefix(lower, "(")) {
		return parseGroup(strings.TrimSpace(s))
	}
	if strings.Contains(lower, "and ") || strings.Contains(lower, " and") || strings.Contains(lower, "or ") || strings.Contains(lower, " or") {
		if strings.HasPrefix(lower, "and") || strings.HasPrefix(lower, "or") {
			return parseLogicalRight(strings.TrimSpace(s))
		}
		return parseLogical(strings.TrimSpace(s))
	}
	return parseAttribute(strings.TrimSpace(s))
}