package mm_utils

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

/*
PagePageSizeToLimitOffset transforms the Page and PageSize parameter in Limit and Offset for DB requests.
*/
func PagePageSizeToLimitOffset(page int, pageSize int) (int, int) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return limit, offset
}

/*
GetOptionalUUIDFromString transforms an optional string in an optional UUID.
*/
func GetOptionalUUIDFromString(input *string) *uuid.UUID {
	var result *uuid.UUID
	if input != nil {
		resultData, err := uuid.Parse(*input)
		if err == nil {
			result = &resultData
		}
	}
	return result
}

/*
GetUUIDFromString transforms a string in a UUID.
*/
func GetUUIDFromString(input string) uuid.UUID {
	return uuid.MustParse(input)
}

/*
GetOptionalStringFromUUID transforms an optional UUID in an optional string.
*/
func GetOptionalStringFromUUID(input *uuid.UUID) *string {
	var result *string

	if input != nil {
		resultConversion := (*input).String()
		result = &resultConversion
	}
	return result
}

/*
GetStringFromUUID transforms a UUID in a string.
*/
func GetStringFromUUID(input uuid.UUID) string {
	return input.String()
}

/*
GetOptionalTimeFromString transforms an optional String in an optional Time.
*/
func GetOptionalTimeFromString(input *string) *time.Time {
	if input != nil {
		parsedTime, err := time.Parse(time.RFC3339, *input)
		if err != nil {
			return nil
		}
		utcParsedTime := parsedTime.UTC()
		return &utcParsedTime
	}
	return nil
}

/*
GetTimeFromString transforms a String in a Time.
*/
func GetTimeFromString(input string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, input)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Invalid Time conversion. Value %s", input), zap.String("service", "utils"), zap.Error(err))
		panic(err)
	}
	return parsedTime.UTC()
}

/*
TransformToStrings transforms a list of inputs into list of strings represented as interfaces.
*/
func TransformToStrings(input []interface{}) []interface{} {
	var items []interface{}
	for _, item := range input {
		items = append(items, fmt.Sprintf("%v", item))
	}
	return items
}

/*
TransformToInterfaces transforms []strings to []interfaces{}.
*/
func TransformToInterfaces(input []string) []interface{} {
	var items []interface{}
	for _, item := range input {
		items = append(items, item)
	}
	return items
}

/*
IsEmpty checks if a value is empty. Are considered empty values: new empty struct, nil, 0, false, "".
*/
func IsEmpty(data any) bool {
	if data != nil {
		return reflect.ValueOf(data).IsZero()
	}
	return true
}

/*
Return a pointer to a boolean
*/
func BoolPtr(b bool) *bool {
	return &b
}

/*
Return a pointer to a float64
*/
func Float64Ptr(b float64) *float64 {
	return &b
}

/*
Return a pointer to a int64
*/
func Int64Ptr(b int64) *int64 {
	return &b
}

/*
Round a Float value pointer to max 2 decimals
*/
func RoundTo2DecimalsPtr(val *float64) *float64 {
	if val == nil {
		return nil
	}
	v := math.Round(*val*100) / 100
	return &v
}

/*
Round a Float value to max 2 decimals
*/
func RoundTo2Decimals(val float64) float64 {
	return math.Round(val*100) / 100
}

/*
DiffStructs returns the list of fields that are different between two structs.
*/
func DiffStructs[T any](current, new T) []string {
	av := reflect.ValueOf(current)
	bv := reflect.ValueOf(new)
	t := reflect.TypeOf(current)

	if av.Kind() != reflect.Struct {
		panic("DiffStructs: only works with struct types")
	}

	var diffs []string
	for i := 0; i < av.NumField(); i++ {
		if !reflect.DeepEqual(av.Field(i).Interface(), bv.Field(i).Interface()) {
			diffs = append(diffs, t.Field(i).Name)
		}
	}
	return diffs
}
