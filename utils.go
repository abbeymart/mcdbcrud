// @Author: abbeymart | Abi Akindele | @Created: 2021-06-24 | @Updated: 2021-06-24
// @Company: mConnect.biz | @License: MIT
// @Description: crud-utility-helper-functions

package mcdbcrud

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/abbeymart/mcresponse"
	"github.com/asaskevich/govalidator"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

type EmailUserNameType struct {
	Email    string
	Username string
}

// EmailUsername processes and returns the loginName as email or username
func EmailUsername(loginName string) EmailUserNameType {
	if govalidator.IsEmail(loginName) {
		return EmailUserNameType{
			Email:    loginName,
			Username: "",
		}
	}

	return EmailUserNameType{
		Email:    "",
		Username: loginName,
	}

}

func TypeOf(rec interface{}) reflect.Type {
	return reflect.TypeOf(rec)
}

// ParseRawValues process the raw rows/records from SQL-query
func ParseRawValues(rawValues [][]byte) ([]interface{}, error) {
	// variables
	var value interface{}
	var values []interface{}
	// parse the current-raw-values
	for _, val := range rawValues {
		if err := json.Unmarshal(val, &value); err != nil {
			return nil, errors.New(fmt.Sprintf("Error parsing raw-row-value: %v", err.Error()))
		} else {
			values = append(values, value)
		}
	}
	return values, nil
}

// ArrayStringContains check if a slice of string contains/includes a string value
func ArrayStringContains(arr []string, val string) bool {
	for _, a := range arr {
		if strings.ToLower(a) == strings.ToLower(val) {
			return true
		}
	}
	return false
}

// ArrayIntContains check if a slice of int contains/includes an int value
func ArrayIntContains(arr []int, val int) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

// ArrayToSQLStringValues transforms a slice of string to SQL-string-formatted-values
func ArrayToSQLStringValues(arr []string) string {
	result := ""
	for ind, val := range arr {
		result += "'" + val + "'"
		if ind < len(arr)-1 {
			result += ", "
		}
	}
	return result
}

// JsonToStruct converts json inputs to equivalent struct data type specification
// rec must be a pointer to a type matching the jsonRec
func JsonToStruct(jsonRec []byte, rec interface{}) error {
	if err := json.Unmarshal(jsonRec, &rec); err == nil {
		return nil
	} else {
		return errors.New(fmt.Sprintf("Error converting json-to-record-format: %v", err.Error()))
	}
}

// DataToValueParam accepts only a struct type/model and returns the ActionParamType
// data camel/Pascal-case keys are converted to underscore-keys to match table-field/columns specs
func DataToValueParam(rec interface{}) (ActionParamType, error) {
	// validate recs as struct{} type
	recType := fmt.Sprintf("%v", reflect.TypeOf(rec).Kind())
	switch recType {
	case "struct":
		dataValue := ActionParamType{}
		v := reflect.ValueOf(rec)
		typeOfS := v.Type()

		for i := 0; i < v.NumField(); i++ {
			dataValue[govalidator.CamelCaseToUnderscore(typeOfS.Field(i).Name)] = v.Field(i).Interface()
			//fmt.Printf("Field: %s\tValue: %v\n", typeOfS.Field(i).ItemName, v.Field(i).Interface())
		}
		return dataValue, nil
	default:
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type struct{}"))
	}
}

// StructToMap function converts struct to map
func StructToMap(rec interface{}) (map[string]interface{}, error) {
	// validate recs as struct{} type
	recType := fmt.Sprintf("%v", reflect.TypeOf(rec).Kind())
	switch recType {
	case "struct":
		break
	default:
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type struct{}"))
	}
	var mapData map[string]interface{}
	// json record
	jsonRec, err := json.Marshal(rec)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing struct to map: %v", err.Error()))
	}
	// json-to-map
	err = json.Unmarshal(jsonRec, &mapData)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing struct to map: %v", err.Error()))
	}
	return mapData, nil
}

// TagField return the field-tag (e.g. table-column-name) for mcorm tag
func TagField(rec interface{}, fieldName string, tag string) (string, error) {
	// validate recs as struct{} type
	t := reflect.TypeOf(rec)
	recType := fmt.Sprintf("%v", t.Kind())
	switch recType {
	case "struct":
		break
	default:
		return "", errors.New(fmt.Sprintf("rec parameter must be of type struct{}"))
	}
	// convert the first-letter to upper-case (public field)
	field, found := t.FieldByName(strings.Title(fieldName))
	if !found {
		// check private field
		field, found = t.FieldByName(fieldName)
		if !found {
			return "", errors.New(fmt.Sprintf("error retrieving tag-field for field-name: %v", fieldName))
		}
	}
	//tagValue := field.Tag
	return field.Tag.Get(tag), nil
}

// StructToTagMap function converts struct to map (tag/underscore_field), for crud-db-table-record
func StructToTagMap(rec interface{}, tag string) (map[string]interface{}, error) {
	// validate recs as struct{} type
	recType := fmt.Sprintf("%v", reflect.TypeOf(rec).Kind())
	switch recType {
	case "struct":
		break
	default:
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type struct{}"))
	}
	tagMapData := map[string]interface{}{}
	mapData, err := StructToMap(rec)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing struct to map: %v", err.Error()))
	}
	// compose tagMapData
	for key, val := range mapData {
		tagField, tagErr := TagField(rec, key, tag)
		if tagErr != nil {
			return nil, errors.New(fmt.Sprintf("error computing tag-field: %v", tagErr.Error()))
		}
		tagMapData[tagField] = val
	}
	return tagMapData, nil
}

func ToCamelCase(text string, sep string) string {
	// accept words/text and separator(' ', '_', '__', '.')
	textArray := strings.Split(text, sep)
	// convert the first word to lowercase
	firstWord := strings.ToLower(textArray[0])
	// convert other words: first letter to upper case and other letters to lowercase
	remWords := textArray[1:]
	var otherWords []string
	for _, item := range remWords {
		// convert first letter to upper case
		item0 := strings.ToUpper(string(item[0]))
		// convert other letters to lowercase
		item1N := strings.ToLower(item[1:])
		itemString := fmt.Sprintf("%v%v", item0, item1N)
		otherWords = append(otherWords, itemString)
	}
	return fmt.Sprintf("%v%v", firstWord, strings.Join(otherWords, ""))
}

// StructToMapUnderscore converts struct to map (underscore_fields), for crud-db-table-record
func StructToMapUnderscore(rec interface{}) (map[string]interface{}, error) {
	// validate recs as struct{} type
	recType := fmt.Sprintf("%v", reflect.TypeOf(rec).Kind())
	switch recType {
	case "struct":
		break
	default:
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type struct{}"))
	}

	caseUnderscoreMapData := map[string]interface{}{}
	mapData, err := StructToMap(rec)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing struct to map: %v", err.Error()))
	}
	// compose caseUnderscoreMapData
	for key, val := range mapData {
		caseUnderscoreMapData[govalidator.CamelCaseToUnderscore(key)] = val
	}
	return caseUnderscoreMapData, nil
}

// MapToMapUnderscore converts map camelCase-fields to underscore-fields
func MapToMapUnderscore(rec interface{}) (map[string]interface{}, error) {
	// validate recs as map type
	recMap, ok := rec.(map[string]interface{})
	if !ok || recMap == nil {
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type map[string]interface{}"))
	}

	uMapData := map[string]interface{}{}
	// compose uMapData
	for key, val := range recMap {
		uMapData[govalidator.CamelCaseToUnderscore(key)] = val
	}
	return uMapData, nil
}

// MapToMapCamelCase converts map underscore-fields to camelCase-fields
func MapToMapCamelCase(rec interface{}, sep string) (map[string]interface{}, error) {
	// validate recs as map type
	recMap, ok := rec.(map[string]interface{})
	if !ok || recMap == nil {
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type map[string]interface{}"))
	}

	uMapData := map[string]interface{}{}
	// compose uMapData
	for key, val := range recMap {
		uMapData[ToCamelCase(key, sep)] = val
	}
	return uMapData, nil
}

// ArrayMapToMapUnderscore converts []map-fields to underscore
func ArrayMapToMapUnderscore(rec interface{}) ([]map[string]interface{}, error) {
	// validate recs as []map type
	arrayMap, ok := rec.([]map[string]interface{})
	if !ok || arrayMap == nil {
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type []map[string]interface{}"))
	}

	var uArrayMapData []map[string]interface{}
	// compose underscoreMapData
	for _, mapRec := range arrayMap {
		uMapData, err := MapToMapUnderscore(mapRec)
		if err != nil {
			return nil, err
		}
		uArrayMapData = append(uArrayMapData, uMapData)
	}

	return uArrayMapData, nil
}

// StructToFieldValues converts struct to record fields(underscore) and associated values (columns and values)
func StructToFieldValues(rec interface{}) ([]string, []interface{}, error) {
	// validate recs as struct{} type
	recType := fmt.Sprintf("%v", reflect.TypeOf(rec).Kind())
	switch recType {
	case "struct":
		break
	default:
		return nil, nil, errors.New(fmt.Sprintf("rec parameter must be of type struct{}"))
	}
	var tableFields []string
	var fieldValues []interface{}
	mapDataValue, err := StructToMap(rec)
	if err != nil {
		return nil, nil, errors.New("error computing struct to map")
	}
	// compose table fields/column(underscore) and values
	for key, val := range mapDataValue {
		tableFields = append(tableFields, govalidator.CamelCaseToUnderscore(key))
		fieldValues = append(fieldValues, val)
	}
	return tableFields, fieldValues, nil
}

// ArrayMapToStruct converts []map/actParams to []struct/model-type
func ArrayMapToStruct(actParams ActionParamsType, recs interface{}) (interface{}, error) {
	// validate recs as slice / []struct{} type
	recsType := fmt.Sprintf("%v", reflect.TypeOf(recs).Kind())
	switch recsType {
	case "slice":
		break
	default:
		return nil, errors.New(fmt.Sprintf("recs parameter must be of type []struct{}: %v", recsType))
	}
	switch rType := recs.(type) {
	case []interface{}:
		for i, val := range rType {
			// validate each record as struct type
			recType := fmt.Sprintf("%v", reflect.TypeOf(val).Kind())
			switch recType {
			case "struct":
				break
			default:
				return nil, errors.New(fmt.Sprintf("recs[%v] parameter must be of type struct{}: %v", i, recType))
			}
		}
	default:
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type []struct{}: %v", rType))
	}
	// compute json records from actParams
	jsonRec, err := json.Marshal(actParams)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing map to struct records: %v", err.Error()))
	}
	// transform json records to []struct{} (recs)
	err = json.Unmarshal(jsonRec, &recs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing map to struct records: %v", err.Error()))
	}
	return recs, nil
}

// MapToStruct converts map to struct
func MapToStruct(mapRecord map[string]interface{}, rec interface{}) (interface{}, error) {
	// validate rec as struct{} type
	recType := fmt.Sprintf("%v", reflect.TypeOf(rec).Kind())
	switch recType {
	case "struct":
		break
	default:
		return nil, errors.New(fmt.Sprintf("rec parameter must be of type struct{}"))
	}
	// compute json records from actParams (map-record)
	jsonRec, err := json.Marshal(mapRecord)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing map to struct records: %v", err.Error()))
	}
	// transform json record to struct{} (rec)
	err = json.Unmarshal(jsonRec, &rec)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error computing map to struct records: %v", err.Error()))
	}
	return rec, nil
}

// GetParamsMessage compose the message-object into mcresponse.ResponseMessage
func GetParamsMessage(msgObject MessageObject) mcresponse.ResponseMessage {
	var messages = ""

	for key, val := range msgObject {
		if messages != "" {
			messages = fmt.Sprintf("%v | %v: %v", messages, key, val)
		} else {
			messages = fmt.Sprintf("%v: %v", key, val)
		}
	}
	return mcresponse.GetResMessage("validateError", mcresponse.ResponseMessageOptions{
		Message: messages,
		Value:   nil,
	})
}

// ConvertJsonStringToMapValue converts the db-json-string-value to the map-type
func ConvertJsonStringToMapValue(jsonStr string) (map[string]interface{}, error) {
	mapVal := map[string]interface{}{}
	jErr := json.Unmarshal([]byte(jsonStr), &mapVal)
	if jErr != nil {
		return nil, jErr
	}
	return mapVal, nil
}

// ConvertJsonStringToTypeValue converts the db-json-string-value to the base-type
func ConvertJsonStringToTypeValue(jsonStr string, typePointer interface{}) (interface{}, error) {
	jErr := json.Unmarshal([]byte(jsonStr), typePointer)
	if jErr != nil {
		return nil, jErr
	}
	return typePointer, nil
}

// ConvertJsonBase64StringToTypeValue converts the db-json-string-value to the base-type
func ConvertJsonBase64StringToTypeValue(base64Str interface{}, typePointer interface{}) (interface{}, error) {
	// assert the base64String value as of string-type
	strVal, ok := base64Str.(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("unable to convert base64-string [%v] to string", base64Str))
	}
	// decode the base64StringValue
	decoded, err := base64.StdEncoding.DecodeString(strVal)
	if err != nil {
		return nil, err
	}
	// transform/un-marshal the decoded value to the base-type
	jErr := json.Unmarshal(decoded, typePointer)
	if jErr != nil {
		return nil, jErr
	}
	return typePointer, nil
}

// ConvertJsonBase64StringToMap converts the db-json-string-value to the map-type
func ConvertJsonBase64StringToMap(base64Str interface{}) (map[string]interface{}, error) {
	mapVal := map[string]interface{}{}
	strVal, ok := base64Str.(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("unable to convert base64-string [%v] to string", base64Str))
	}
	decoded, err := base64.StdEncoding.DecodeString(strVal)
	if err != nil {
		return nil, err
	}
	jErr := json.Unmarshal(decoded, &mapVal)
	if jErr != nil {
		return nil, jErr
	}
	return mapVal, nil
}

func ConvertByteSliceToBase64Str(fileContent []byte) string {
	return base64.StdEncoding.EncodeToString(fileContent)
}

func ConvertStringToBase64Str(fileContent string) string {
	return base64.StdEncoding.EncodeToString([]byte(fileContent))
}

func ExcludeEmptyIdFromMapRecord(rec ActionParamType) ActionParamType {
	mapVal := ActionParamType{}
	for key, val := range rec {
		if key == "id" && val == "" {
			continue
		}
		mapVal[key] = val
	}
	return mapVal
}

// ExcludeFieldFromMapRecord exclude id and accessKey fields
func ExcludeFieldFromMapRecord(rec ActionParamType, field string) ActionParamType {
	mapVal := ActionParamType{}
	for key, val := range rec {
		if key == field {
			continue
		}
		mapVal[key] = val
	}
	return mapVal
}

func ExcludeEmptyIdFields(recs []ActionParamType) []ActionParamType {
	var mapValues []ActionParamType
	for _, rec := range recs {
		mapVal := ActionParamType{}
		for key, val := range rec {
			if (key == "id" || strings.HasSuffix(key, "Id")) && (val == nil || val == "") {
				continue
			}
			mapVal[key] = val
		}
		mapValues = append(mapValues, mapVal)
	}
	return mapValues
}

func StructToMapToCamelCase(rec interface{}, sep string) (map[string]interface{}, error) {
	mapVal, mErr := StructToMap(rec)
	if mErr != nil {
		return nil, mErr
	}
	val, err := MapToMapCamelCase(mapVal, sep)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ComputeTaskDuration computes the task interval in microseconds
func ComputeTaskDuration(start time.Time, end time.Time) int64 {
	return end.Sub(start).Microseconds()
}

// RandomString generates random string of characters and numbers
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// RandomNumbers generates random numbers using rand.Perm and returns []int as string
func RandomNumbers(n int) string {
	rand.Seed(time.Now().UnixNano())
	v := rand.Perm(n)
	var vString []string
	for _, item := range v {
		vString = append(vString, fmt.Sprintf("%v", item))
	}
	return fmt.Sprintf("%v", strings.Join(vString, ""))
}

// CheckTaskTypeV1 determine/set task type based on the actionParams - update tag-v0.6.2 to v0.6.3
func CheckTaskTypeV1(params CrudParamsType) string {
	taskType := ""
	if len(params.ActionParams) > 0 {
		actParam := params.ActionParams[0]
		_, ok := actParam["id"]
		if !ok {
			if len(params.RecordIds) > 0 || len(params.QueryParams) > 0 {
				taskType = UpdateTask
			} else {
				taskType = CreateTask
			}
		} else {
			taskType = UpdateTask
		}
	}
	return taskType
}

// CheckTaskType function determines and returns the taskType, based on the actionParams from the CrudParamsType
func CheckTaskType(params CrudParamsType) string {
	if len(params.ActionParams) < 1 {
		return UnknownTask
	}
	// check task-types for actionParams === 1 or > 1
	if len(params.ActionParams) == 1 {
		actParam := params.ActionParams[0]
		recId, ok := actParam["id"].(string)
		if !ok || recId == "" {
			if len(params.RecordIds) > 0 || len(params.QueryParams) > 0 {
				return UpdateTask
			} else {
				return CreateTask
			}
		} else {
			return UpdateTask
		}
	}
	if len(params.ActionParams) > 1 {
		updateCount := 0
		createCount := 0
		for _, actRec := range params.ActionParams {
			recId, ok := actRec["id"].(string)
			if !ok || recId == "" {
				createCount += 1
			} else {
				updateCount += 1
			}
		}
		// determine task-type
		if updateCount > 0 && createCount > 0 {
			return UnknownTask
		}
		if createCount > 0 {
			return CreateTask
		}
		if updateCount > 0 {
			return UpdateTask
		}
	}
	return UnknownTask
}

// ComputeAccessResValue returns the transform access-response-value of AccessResValueType
func ComputeAccessResValue(params mcresponse.ResponseMessage) AccessResValueType {
	pJson, _ := json.Marshal(params.Value)
	var accessResValue AccessResValueType
	err := json.Unmarshal(pJson, &accessResValue)
	if err != nil {
		return AccessResValueType{}
	}
	return accessResValue
}

// ValidateSubActionParams validates that subscriber-appIds includes actionParam-appId, for save - create/update tasks
func ValidateSubActionParams(actParams ActionParamsType, subAppIds []string) bool {
	result := false
	for _, rec := range actParams {
		id, idOk := rec["appId"].(string)
		if !idOk || !ArrayStringContains(subAppIds, id) {
			return false
		}
		result = true
	}
	return result
}

// TransformGetCrudParams compose the crud-params for read-query
func TransformGetCrudParams(params CrudParamsType, accessRes mcresponse.ResponseMessage) CrudParamsType {
	isAdmin := false
	systemAdmin := false
	userAdmin := false
	var subAppIds []string
	// check the subscriber-admin-access
	if accessRes.Code == "success" {
		subAdminValue, subAdminOk := accessRes.Value.(AccessResValueType)
		if subAdminOk {
			isAdmin = subAdminValue.IsAdmin
			systemAdmin = subAdminValue.SystemAdmin
			userAdmin = subAdminValue.UserAdmin
			subAppIds = subAdminValue.AppIds
		}
	}
	// if admin, return params, as-is
	if isAdmin || systemAdmin || userAdmin {
		return params
	}
	// if subscriber has active applications, as admin, query by subscriber-appIds
	if len(subAppIds) > 0 {
		params.RecordIds = []string{}
		params.QueryParams = map[string]interface{}{
			"appId": subAppIds,
		}
		return params
	}
	// otherwise, apply to record(s) createdBy the current user/userId only
	params.RecordIds = []string{}
	params.QueryParams = map[string]interface{}{
		"createdBy": params.UserInfo.UserId,
	}
	return params
}

// QueryFields function computes the underscore field-names from the specified model
func QueryFields(modelRef interface{}) (string, error) {
	// compute map[string]interface (underscore_fields) from the modelRef (struct)
	mapMod, mapErr := StructToMapUnderscore(modelRef)
	if mapErr != nil {
		return "", mapErr
	}
	// compute table-fields
	var fieldNames []string
	for fieldName := range mapMod {
		fieldNames = append(fieldNames, fieldName)
	}
	fieldLen := len(fieldNames)
	fieldText := ""
	for i, fieldName := range fieldNames {
		//fieldText += "'" + fieldName + "'"
		fieldText += fieldName
		if i < fieldLen-1 {
			fieldText += ", "
		}
	}
	return fieldText, nil
}
