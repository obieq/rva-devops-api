package dynamodb

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	aws "github.com/AdRoll/goamz/aws"
	dynamo "github.com/AdRoll/goamz/dynamodb"
	"github.com/imdario/mergo"
	"github.com/joho/godotenv"
	. "github.com/obieq/goar"
)

const DB_PRIMARY_KEY_NAME string = "id"
const MODEL_PRIMARY_KEY_NAME string = "ID"

type ArDynamodb struct {
	ActiveRecord
	ID string `json:"id,omitempty"`
	Timestamps
}

var (
	client *dynamo.Server
)

var envVars = func() (map[string]string, error) {
	return godotenv.Read()
}

var connectOpts = func() map[string]string {
	opts := make(map[string]string)

	if envs, err := envVars(); err != nil {
		panic(fmt.Errorf("Error loading .env file: %v", err))
	} else {
		opts["accessKey"] = os.Getenv(envs["AWS_ACCESS_KEY_ID"])
		opts["secretKey"] = os.Getenv(envs["AWS_SECRET_ACCESS_KEY"])
	}

	return opts
}

func connect() *dynamo.Server {
	opts := connectOpts()

	region := aws.USEast
	auth := aws.Auth{AccessKey: opts["accessKey"], SecretKey: opts["secretKey"]}

	return dynamo.New(auth, region)
}

func init() {
	client = connect()
}

func Client() *dynamo.Server {
	return client
}

func (ar *ArDynamodb) SetKey(key string) {
	ar.ID = key
}

func (ar *ArDynamodb) All(models interface{}, opts map[string]interface{}) (err error) {
	return errors.New("All method not supported by Dynamodb.  Create a View instead.")
}

func (ar *ArDynamodb) Truncate() (numRowsDeleted int, err error) {
	return -1, errors.New("Truncate method not yet implemented")
}

func (ar *ArDynamodb) Find(key interface{}) (interface{}, error) {
	self := ar.Self()
	tbl, dynamoKey := ar.GetTableWithPrimaryKey(key)

	err := tbl.GetDocument(dynamoKey, self)

	// NOTE: the AdRoll sdk returns an error is the key doesn't exist
	if err == nil {
		// set the ID b/c the AdRoll sdk purposefully empties it for some reason
		pointer := reflect.Indirect(reflect.ValueOf(self))
		field := pointer.FieldByName(MODEL_PRIMARY_KEY_NAME)
		field.SetString(key.(string))
	} else {
		self = nil
	}

	return self, err
}

func (ar *ArDynamodb) DbSave() error {
	tbl, key := ar.GetTableWithPrimaryKey(ar.ID)
	return tbl.PutDocument(key, ar.Self())
}

func (ar *ArDynamodb) Patch() (bool, error) {
	var err error
	var success bool = false
	tbl, key := ar.GetTableWithPrimaryKey(ar.ID)

	// copy self for later use if we perorm an update
	// reason being that when find is called, it will update
	// the underlying Self() instance
	source := reflect.ValueOf(ar.Self()).Elem().Interface()

	// query the db to determine if we're doing an insert or an update
	// NOTE: due to the fact this supports PATCH updates, we need to
	//       get the persisted instance if one exists in order to update
	//       the subset of fields.  If we didn't do so, then biz rule validations
	//       could fail b/c of the incomplete data
	dbInstance, err := ar.Find(ar.ID)

	if err == nil { // instance found, so update
		var arr []reflect.Value

		// sync db instance and self instance
		e := reflect.ValueOf(dbInstance).Elem()
		addr := e.Addr()
		method := addr.MethodByName("Self")
		destination := method.Call(arr)[0].Interface()

		// merge patch changes into existing instance
		// NOTE: nil/empty values don't appear to overwrite existing values
		err = mergo.Merge(destination, source)

		if err == nil {
			// set updated at timestamp
			updatedAt := time.Now().UTC()
			ar.UpdatedAt = &updatedAt // given the use of pointers, no need to use reflection

			// run validations and update
			success = ar.Valid()
			if success == true {
				err = tbl.PutDocument(key, destination)
			}
		}
	}

	return success, err
}

func (ar *ArDynamodb) DbDelete() (err error) {
	primary := dynamo.NewStringAttribute(DB_PRIMARY_KEY_NAME, "")
	pk := dynamo.PrimaryKey{KeyAttribute: primary}
	t := dynamo.Table{Server: client, Name: ar.ModelName(), Key: pk}

	dynamoKey := &dynamo.Key{HashKey: ar.ID}
	return t.DeleteDocument(dynamoKey)
}

func (ar *ArDynamodb) DbSearch(models interface{}) (err error) {
	return errors.New("Search method not supported by Dynamodb.  Create a View instead.")
}

func (ar *ArDynamodb) GetTableWithPrimaryKey(key interface{}) (dynamo.Table, *dynamo.Key) {
	// primary key initialization example
	//     https://github.com/AdRoll/goamz/blob/c73835dc8fc6958baf8df8656864ee4d6d04b130/dynamodb/query_builder_test.go
	//         primary := NewStringAttribute("TestHashKey", "")
	//         secondary := NewNumericAttribute("TestRangeKey", "")
	//         key := PrimaryKey{primary, secondary}
	primary := dynamo.NewStringAttribute(DB_PRIMARY_KEY_NAME, "")
	pk := dynamo.PrimaryKey{KeyAttribute: primary}
	t := dynamo.Table{Server: client, Name: ar.ModelName(), Key: pk}
	dynamoKey := &dynamo.Key{HashKey: key.(string)}

	return t, dynamoKey
}
