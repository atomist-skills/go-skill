package skills

import "fmt"

// ParseMultiChoiceArg parses the multi-choice skill parameter into a string array.
func ParseMultiChoiceArg(arg interface{}) []string {
	var parsedArgs []string

	if arg == nil {
		return parsedArgs
	}

	if array, ok := arg.([]interface{}); ok {
		for _, arg := range array {
			parsedArgs = append(parsedArgs, fmt.Sprintf("%v", arg))
		}
	}

	return parsedArgs
}

// ParseStringArrayArgs parses the string-array skill parameter into a string array.
func ParseStringArrayArg(arg interface{}) []string {
	var parsedArgs []string

	if arg == nil {
		return parsedArgs
	}

	if array, ok := arg.([]interface{}); ok {
		for _, arg := range array {
			parsedArgs = append(parsedArgs, fmt.Sprintf("%v", arg))
		}
	}

	return parsedArgs
}

// ParseIntArg parses the int skill parameter into an int64.
func ParseIntArg(arg interface{}) int64 {
	if arg == nil {
		return 0
	}

	parsedArgAsInt64, ok := arg.(int64)
	if ok {
		return parsedArgAsInt64
	}

	parsedArgAsFloat, ok := arg.(float64)
	if ok {
		return int64(parsedArgAsFloat)
	}

	return 0
}
