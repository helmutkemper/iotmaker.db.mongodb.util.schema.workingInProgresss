package iotmakerdbmongodbutilschema

import (
	"errors"
	"reflect"
)

// The object schema type configures the content of documents.
// For more information, see the official JSON Schema object guide.
// https://json-schema.org/understanding-json-schema/reference/object.html
//
//   Example:
//   {
//     "bsonType": "object",
//     "title": "<Type Name>",
//     "required": ["<Required Field Name>", ...],
//     "properties": {
//       "<Field Name>": <Schema Document>
//     },
//     "minProperties": <integer>,
//     "maxProperties": <integer>,
//     "patternProperties": {
//       "<Field Name Regex>": <Schema Document>
//     },
//     "additionalProperties": <boolean> | <Schema Document>,
//     "dependencies": {
//       "<Field Name>": <Schema Document> | ["<Field Name>", ...]
//     }
//   }
type TypeBsonObject struct {
	TypeBsonCommonToAllTypes

	// An object where each field maps to a field in the parent object by name. The value
	// of each field is a schema document that configures the value of the field.
	Properties map[string]map[string]BsonType

	// The minimum number of fields allowed in the object.
	MinPropertiesHasSet bool
	MinProperties       int64

	// The maximum number of fields allowed in the object.
	MaxPropertiesHasSet bool
	MaxProperties       int64

	// An object where each field is a regular expression string that maps to all fields in
	// the parent object that match. The value of each field is a schema document that
	// configures the value of matched fields.
	PatternProperties []PatternProperties

	// Default: true.
	// If true, a document may contain additional fields that are not defined in the
	// schema.
	// If false, only fields that are explicitly defined in the schema may appear in a
	// document.
	// If the value is a schema object, any additional fields must validate against the
	// schema
	AdditionalPropertiesBoolIsSet bool
	AdditionalPropertiesBoolValue bool
	AdditionalPropertiesMap       map[string]map[string]BsonType

	// Specify property and schema dependencies.
	// https://www.mongodb.com/blog/post/json-schema-validation--dependencies-you-can-depend-on
	Dependencies map[string]map[string]BsonType

	Required map[string]bool
}

func (el *TypeBsonObject) getTypeString() string {
	return "object"
}

func (el *TypeBsonObject) Populate(schema map[string]interface{}) (err error) {
	err = el.populateGeneric(schema)
	if err != nil {
		return
	}

	el.MinPropertiesHasSet, el.MinProperties, err = el.getPropertyMinProperties(schema)
	if err != nil {
		return
	}

	el.MaxPropertiesHasSet, el.MaxProperties, err = el.getPropertyMaxProperties(schema)
	if err != nil {
		return
	}

	err = el.populateRequired(schema)
	if err != nil {
		return
	}

	el.Properties, err = el.populateBsonType(schema)
	return
}

// processRequiredFields (English): Process the required fields
//    json schema example:
//    {
//      "bsonType": "object",
//      "title": "<Type Name>",
//      "required": ["<Required Field Name>", ...],
//      "properties": {
//        "<Field Name>": <Schema Document>
//      }
//    }
//
// processRequiredFields (Português): Processa os campos requeridos
//    exemplo de esquema json:
//    {
//      "bsonType": "object",
//      "title": "<Type Name>",
//      "required": ["<Required Field Name>", ...],
//      "properties": {
//        "<Field Name>": <Schema Document>
//      }
//    }
func (el *TypeBsonObject) populateRequired(schema map[string]interface{}) (err error) {
	return el.populateRequiredSupport(&el.Required, "", schema)
}

func (el *TypeBsonObject) populateRequiredSupport(requiredPointer *map[string]bool, key string, schema map[string]interface{}) (err error) {
	var found bool
	//var newSchema map[string]interface{}

	var requiredList []interface{}
	requiredList, found = schema["required"].([]interface{})
	if found == false {
		return
	}

	if *requiredPointer == nil {
		*requiredPointer = make(map[string]bool)
	}

	for _, requiredKeyName := range requiredList {
		if key != "" {
			//todo: verificar string
			requiredKeyName = key + "." + requiredKeyName.(string)
		}
		(*requiredPointer)[requiredKeyName.(string)] = true
	}

	// fixme: início: isto está correto, mas, o objeto só deve verificar aos campos dele ou entrar na árvore e verificar?
	//newSchema, _ = schema["properties"].(map[string]interface{})
	//for schemaCellKey, schemaCell := range newSchema {
	//  if key != "" {
	//    schemaCellKey = key + "." + schemaCellKey
	//  }
	//  err = el.populateRequiredSupport(requiredPointer, schemaCellKey, schemaCell.(map[string]interface{}))
	//}
	// fixme: fim

	return
}

func (el *TypeBsonObject) Verify(value interface{}) (err error) {

	err = el.verifyMaxProperties()
	if err != nil {
		return
	}

	err = el.verifyMinProperties()
	return
}

func (el *TypeBsonObject) getPropertyMinProperties(schema map[string]interface{}) (set bool, minimum int64, err error) {
	var found bool

	_, found = schema["minProperties"]
	if found == false {
		return
	}

	set = true
	minimum, err = el.getPropertyAsInt64(schema, "minProperties")
	return
}

func (el *TypeBsonObject) getPropertyMaxProperties(schema map[string]interface{}) (set bool, minimum int64, err error) {
	var found bool

	_, found = schema["maxProperties"]
	if found == false {
		return
	}

	set = true
	minimum, err = el.getPropertyAsInt64(schema, "maxProperties")
	return
}

func (el *TypeBsonObject) populateBsonType(schema map[string]interface{}) (properties map[string]map[string]BsonType, err error) {

	properties = make(map[string]map[string]BsonType)
	//var typeList []string
	//typeList, err = el.getPropertyBsonTypeAsSlice(schema)
	//for _, v := range typeList {
	//  err = el.typeStringToTypeObjectPopulated(&properties, "", v, schema)
	//  if err != nil {
	//    return
	//  }
	//}

	var newSchema map[string]interface{}
	newSchema, _ = schema["properties"].(map[string]interface{})
	for schemaCellKey, schemaCell := range newSchema {

		var typesInCell []string
		typesInCell, err = el.getPropertyBsonTypeAsSlice(schemaCell.(map[string]interface{}))

		for _, currentType := range typesInCell {
			//if key != "" {
			//  schemaCellKey = key + "." + schemaCellKey
			//}
			err = el.typeStringToTypeObjectPopulated(&properties, schemaCellKey, currentType, schemaCell.(map[string]interface{}))
			if err != nil {
				return
			}
		}
	}

	return
}

func (el *TypeBsonObject) getPropertyBsonTypeAsSlice(schema map[string]interface{}) (value []string, err error) {

	value = make([]string, 0)

	var bsonType interface{}
	var found bool

	bsonType, found = schema["bsonType"]
	if found == false {
		return
	}

	if reflect.ValueOf(bsonType).Kind() == reflect.Slice {
		for _, v := range bsonType.([]interface{}) {
			if reflect.ValueOf(v).Kind() != reflect.String {
				err = errors.New("the 'bsonType' values must be a string")
				return
			}

			value = append(value, v.(string))
		}
		return
	}

	if reflect.ValueOf(bsonType).Kind() == reflect.String {
		value = append(value, bsonType.(string))
		return
	}

	err = errors.New("the 'bsonType' a string or a array of string")
	return
}

func (el *TypeBsonObject) typeStringToTypeObjectPopulated(propertiesPointer *map[string]map[string]BsonType, key string, typeString string, schema map[string]interface{}) (err error) {
	//var newSchema map[string]interface{}
	var objType InterfaceBson

	if *propertiesPointer == nil {
		*propertiesPointer = make(map[string]map[string]BsonType)
	}

	switch typeString {

	// English:
	// in case of enum, 'bsonType' can be omitted.
	// when this happens, the 'TypeBsonCommonToAllTypes' object meets the needs
	// note: 'generic' was created by me and is not provided in the documentation
	//
	// Português:
	// em caso de enum, 'bsonType' pode ser omitido.
	// quando isto acontece, o objeto 'TypeBsonCommonToAllTypes' atende as necessidades
	// nota: 'generic' foi criado por mim e não é previsto na documentação
	case "generic":
		objType = &TypeBsonGeneric{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	case "object":
		objType = &TypeBsonObject{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

		//newSchema, _ = schema["properties"].(map[string]interface{})
		//for schemaCellKey, schemaCell := range newSchema {
		//
		//  var typesInCell []string
		//  typesInCell, err = el.getPropertyBsonTypeAsSlice(schemaCell.(map[string]interface{}))
		//
		//  for _, currentType := range typesInCell {
		//    if key != "" {
		//      schemaCellKey = key + "." + schemaCellKey
		//    }
		//    err = el.typeStringToTypeObjectPopulated(propertiesPointer, schemaCellKey, currentType, schemaCell.(map[string]interface{}))
		//    if err != nil {
		//      return
		//    }
		//  }
		//}

	case "double":
		objType = &TypeBsonDouble{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	case "string":
		objType = &TypeBsonString{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	case "array":
		objType = &TypeBsonArray{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	//case "binData":
	//case "objectId":
	case "bool":
		objType = &TypeBsonBool{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	//case "date":
	//case "null":
	//case "regex":
	//case "dbPointer":
	//case "javascript":
	//case "symbol":
	//case "javascriptWithScope":
	case "int":
		objType = &TypeBsonInt{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	case "timestamp":
	case "long":
		objType = &TypeBsonLong{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	case "decimal":
		objType = &TypeBsonDecimal{}
		err = objType.Populate(schema)
		if err != nil {
			return
		}

	default:
		err = errors.New("type not implemented yet")
	}

	if (*propertiesPointer)[key] == nil {
		(*propertiesPointer)[key] = make(map[string]BsonType)
	}

	(*propertiesPointer)[key][typeString] = BsonType{ElementType: objType}

	return
}
