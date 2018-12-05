package gosql

import (
	"fmt"
	"reflect"
	"strings"
)

// RelationOne is get the associated relational data for a single piece of data
func RelationOne(data interface{}) error {
	refVal := reflect.Indirect(reflect.ValueOf(data))
	t := refVal.Type()
	for i := 0; i < t.NumField(); i++ {
		val := t.Field(i).Tag.Get("relation")
		name := t.Field(i).Name
		field := t.Field(i)

		if val != "" && val != "-" {
			relations := strings.Split(val, ",")

			var foreignModel reflect.Value
			// if field type is slice then one-to-many ,eg: []*Struct
			if field.Type.Kind() == reflect.Slice {
				foreignModel = reflect.New(field.Type)
				// batch get field values
				// Since the structure is slice, there is no need to new Value
				err := Model(foreignModel.Interface()).Where(fmt.Sprintf("%s=?", relations[1]), mapper.FieldByName(refVal, relations[0]).Interface()).All()
				if err != nil {
					return err
				}

				if reflect.Indirect(foreignModel).Len() == 0 {
					// If relation data is empty, must set empty slice
					// Otherwise, the JSON result will be null instead of []
					refVal.FieldByName(name).Set(reflect.MakeSlice(field.Type, 0, 0))
				} else {
					refVal.FieldByName(name).Set(foreignModel.Elem())
				}

			} else {
				// If field type is struct the one-to-one,eg: *Struct
				foreignModel = reflect.New(field.Type.Elem())
				err := Model(foreignModel.Interface()).Where(fmt.Sprintf("%s=?", relations[1]), mapper.FieldByName(refVal, relations[0]).Interface()).Get()
				if err != nil {
					return err
				}

				refVal.FieldByName(name).Set(foreignModel)
			}
		}
	}
	return nil
}

// RelationAll is gets the associated relational data for multiple pieces of data
func RelationAll(data interface{}) error {
	refVal := reflect.Indirect(reflect.ValueOf(data))

	l := refVal.Len()
	// get the struct field in slice
	t := refVal.Index(0).Elem().Type()

	for i := 0; i < t.NumField(); i++ {
		relVals := make([]interface{}, 0)
		val := t.Field(i).Tag.Get("relation")
		name := t.Field(i).Name
		field := t.Field(i)
		if val != "" && val != "-" {
			relations := strings.Split(val, ",")
			// get relation field values
			for j := 0; j < l; j++ {
				relVals = append(relVals, mapper.FieldByName(refVal.Index(j), relations[0]).Interface())
			}

			var foreignModel reflect.Value
			// if field type is slice then one to many ,eg: []*Struct
			if field.Type.Kind() == reflect.Slice {
				foreignModel = reflect.New(field.Type)
				// batch get field values
				// Since the structure is slice, there is no need to new Value
				err := Model(foreignModel.Interface()).Where(fmt.Sprintf("%s in(?)", relations[1]), relVals).All()
				if err != nil {
					return err
				}

				fmap := make(map[interface{}]reflect.Value)

				// Combine relation data as a one-to-many relation
				// For example, if there are multiple images under an article
				// we use the article ID to associate the images, map[1][]*Images
				for n := 0; n < reflect.Indirect(foreignModel).Len(); n++ {
					fid := mapper.FieldByName(refVal.Index(n), relations[0])
					fmap[fid.Interface()] = reflect.New(reflect.SliceOf(field.Type.Elem())).Elem()
					fmap[fid.Interface()] = reflect.Append(fmap[fid.Interface()], reflect.Indirect(foreignModel).Index(n))
				}

				// Set the result to the model
				for j := 0; j < l; j++ {
					fid := mapper.FieldByName(refVal.Index(j), relations[0])
					if value, has := fmap[fid.Interface()]; has {
						reflect.Indirect(refVal.Index(j)).FieldByName(name).Set(value)
					} else {
						// If relation data is empty, must set empty slice
						// Otherwise, the JSON result will be null instead of []
						reflect.Indirect(refVal.Index(j)).FieldByName(name).Set(reflect.MakeSlice(field.Type, 0, 0))
					}
				}
			} else {
				// If field type is struct the one to one,eg: *Struct
				foreignModel = reflect.New(field.Type.Elem())
				// Batch get field values, but must new slice []*Struct
				fi := reflect.New(reflect.SliceOf(foreignModel.Type()))
				err := Model(fi.Interface()).Where(fmt.Sprintf("%s in(?)", relations[1]), relVals).All()
				if err != nil {
					return err
				}

				// Combine relation data as a one-to-one relation
				fmap := make(map[interface{}]reflect.Value)
				for n := 0; n < reflect.Indirect(fi).Len(); n++ {
					fmap[mapper.FieldByName(refVal.Index(n), relations[0]).Interface()] = reflect.Indirect(fi).Index(n)
				}

				// Set the result to the model
				for j := 0; j < l; j++ {
					fid := mapper.FieldByName(refVal.Index(j), relations[0])
					if value, has := fmap[fid.Interface()]; has {
						reflect.Indirect(refVal.Index(j)).FieldByName(name).Set(value)
					}
				}
			}
		}
	}
	return nil
}
