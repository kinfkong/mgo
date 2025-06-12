// modern_utils.go - Utility functions for modern MongoDB driver compatibility wrapper

package mgo

import (
	"reflect"

	"github.com/globalsign/mgo/bson"
	officialBson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversion helpers
func convertMGOToOfficial(input interface{}) interface{} {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case bson.M:
		result := officialBson.M{}
		for key, value := range v {
			result[key] = convertMGOToOfficial(value)
		}
		return result
	case bson.D:
		result := officialBson.D{}
		for _, elem := range v {
			result = append(result, officialBson.E{
				Key:   elem.Name,
				Value: convertMGOToOfficial(elem.Value),
			})
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertMGOToOfficial(item)
		}
		return result
	case bson.ObjectId:
		if len(v) == 12 {
			objID := primitive.ObjectID{}
			copy(objID[:], []byte(v))
			return objID
		}
		return v
	default:
		return v
	}
}

func convertOfficialToMGO(input interface{}) interface{} {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case officialBson.M:
		result := bson.M{}
		for key, value := range v {
			result[key] = convertOfficialToMGO(value)
		}
		return result
	case officialBson.D:
		result := bson.D{}
		for _, elem := range v {
			result = append(result, bson.DocElem{
				Name:  elem.Key,
				Value: convertOfficialToMGO(elem.Value),
			})
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertOfficialToMGO(item)
		}
		return result
	case primitive.ObjectID:
		return bson.ObjectId(v[:])
	default:
		return v
	}
}

// convertSliceWithReflect converts a slice of interfaces to a target slice type using reflection
func convertSliceWithReflect(srcSlice []interface{}, dst interface{}) error {
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return ErrNotFound
	}

	dstSlice := dstValue.Elem()
	if dstSlice.Kind() != reflect.Slice {
		return ErrNotFound
	}

	elementType := dstSlice.Type().Elem()
	newSlice := reflect.MakeSlice(dstSlice.Type(), 0, len(srcSlice))

	for _, item := range srcSlice {
		// Convert each item to the target element type
		newElement := reflect.New(elementType).Interface()
		err := mapStructToInterface(item, newElement)
		if err != nil {
			return err
		}
		newSlice = reflect.Append(newSlice, reflect.ValueOf(newElement).Elem())
	}

	dstSlice.Set(newSlice)
	return nil
}

func mapStructToInterface(src, dst interface{}) error {
	if src == nil {
		return ErrNotFound
	}

	// Handle slice conversion specifically
	if srcSlice, ok := src.([]interface{}); ok {
		// Use reflection to handle slice conversion properly
		return convertSliceWithReflect(srcSlice, dst)
	}

	// Handle single document conversion
	data, err := bson.Marshal(src)
	if err != nil {
		return err
	}
	return bson.Unmarshal(data, dst)
}
