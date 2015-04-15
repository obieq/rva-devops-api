package orchestrate

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/joho/godotenv"
	. "github.com/obieq/goar"
	c "github.com/orchestrate-io/gorc"
)

type ArOrchestrate struct {
	ActiveRecord
	ID string `json:"id,omitempty"`
	Timestamps
}

var (
	client *c.Client
)

var connectOpts = func() map[string]string {
	opts := make(map[string]string)
	if envs, err := godotenv.Read(); err != nil {
		log.Fatal("Error loading .env file")
	} else {
		opts["api_key"] = os.Getenv(envs["ORCHESTRATE_API_KEY"])
	}

	return opts
}

func connect() (client *c.Client) {
	opts := connectOpts()

	return c.NewClient(opts["api_key"])
}

func init() {
	client = connect()
}

func Client() *c.Client {
	return client
}

func (ar *ArOrchestrate) SetKey(key string) {
	ar.ID = key
}

func (ar *ArOrchestrate) All(models interface{}, opts map[string]interface{}) (err error) {
	var limit int = 10 // per Orchestrate's documentation: 10 default, 100 max
	var response *c.KVResults

	// set limit
	if opts["limit"] != nil {
		limit = opts["limit"].(int)
		if limit > 100 { // max limit is 100
			return errors.New("limit must be less than 100")
		}
	}

	// parse options to determine which query to use
	if opts["afterKey"] != nil {
		response, err = client.ListAfter(ar.ModelName(), opts["afterKey"].(string), limit)
	} else if opts["startKey"] != nil {
		response, err = client.ListStart(ar.ModelName(), opts["startKey"].(string), limit)
	} else {
		response, err = client.List(ar.ModelName(), limit)
	}

	if err != nil {
		return err
	}

	return mapResults(response.Results, models)
}

func (ar *ArOrchestrate) Truncate() (numRowsDeleted int, err error) {
	err = client.DeleteCollection(ar.ModelName())

	return -1, err
}

func (ar *ArOrchestrate) Find(id interface{}) (interface{}, error) {
	self := ar.Self()
	modelVal := reflect.ValueOf(self).Elem()
	modelInterface := reflect.New(modelVal.Type()).Interface()

	result, err := client.Get(ar.ModelName(), id.(string))

	if result != nil {
		err = result.Value(&modelInterface)
	} else {
		modelInterface = nil
	}

	return modelInterface, err
}

func (ar *ArOrchestrate) DbSave() error {
	var err error

	if ar.UpdatedAt != nil {
		_, err = client.Put(ar.ModelName(), ar.ID, ar.Self())
	} else {
		_, err = client.PutIfAbsent(ar.ModelName(), ar.ID, ar.Self())
	}

	return err
}

func (ar *ArOrchestrate) DbDelete() (err error) {
	return client.Purge(ar.ModelName(), ar.ID)
}

func (ar *ArOrchestrate) DbSearch(models interface{}) (err error) {
	var query, sort string
	var response *c.SearchResults
	//query := r.Db(DbName()).Table(ar.Self().ModelName())

	// plucks
	//query = processPlucks(query, ar)

	// where conditions
	if query, err = processWhereConditions(ar); err != nil {
		return err
	}

	// aggregations
	//if query, err = processAggregations(query, ar); err != nil {
	//return err
	//}

	// order bys
	sort = processSorts(ar)

	// TODO: delete!
	log.Printf("DbSearch query: %s", query)

	// run search
	if sort == "" {
		if response, err = client.Search(ar.ModelName(), query, 100, 0); err != nil {
			return err
		}
	} else {
		if response, err = client.SearchSorted(ar.ModelName(), query, sort, 100, 0); err != nil {
			return err
		}
	}

	return mapResults(response.Results, models)
}

//func processPlucks(query r.Term, ar *ArRethinkDb) r.Term {
//if plucks := ar.Query().Plucks; plucks != nil {
//query = query.Pluck(plucks...)
//}

//return query
//}

func mapResults(orchestrateResults interface{}, models interface{}) (err error) {
	// now, map orchstrate's raw json to the desired active record type
	modelsv := reflect.ValueOf(models)
	if modelsv.Kind() != reflect.Ptr || modelsv.Elem().Kind() != reflect.Slice {
		panic("models argument must be a slice address")
	}
	slicev := modelsv.Elem()
	elemt := slicev.Type().Elem()

	switch t := orchestrateResults.(type) {
	case []c.KVResult:
		for _, result := range t {
			elemp := reflect.New(elemt)
			if err = result.Value(elemp.Interface()); err != nil {
				return err
			}

			slicev = reflect.Append(slicev, elemp.Elem())
		}
	case []c.SearchResult:
		for _, result := range t {
			elemp := reflect.New(elemt)
			if err = result.Value(elemp.Interface()); err != nil {
				return err
			}

			slicev = reflect.Append(slicev, elemp.Elem())
		}
	default:
		return errors.New(fmt.Sprintf("Orchestrate Response Type Not Mapped: %v", t))
	}

	// assign mapped results to the caller's supplied array
	modelsv.Elem().Set(slicev)

	return err
}

func processWhereConditions(ar *ArOrchestrate) (query string, err error) {
	var whereStmt, whereCondition string

	if len(ar.Query().WhereConditions) > 0 {
		for index, where := range ar.Query().WhereConditions {
			switch where.RelationalOperator {
			case EQ: // equal
				whereCondition = where.Key + ":" + fmt.Sprintf("%v", where.Value)
				//whereCondition = where.Key + ":" + where.Value.(string)
				//whereCondition = r.Row.Field(where.Key).Eq(where.Value)
			//case NE: // not equal
			//whereCondition = r.Row.Field(where.Key).Ne(where.Value)
			//case LT: // less than
			//whereCondition = r.Row.Field(where.Key).Lt(where.Value)
			//case LTE: // less than or equal
			//whereCondition = r.Row.Field(where.Key).Le(where.Value)
			//case GT: // greater than
			//// TODO: create function to set range based on type???
			//whereCondition = where.Key + ":[" + fmt.Sprintf("%v", where.Value) + " TO *]"
			//whereCondition = r.Row.Field(where.Key).Gt(where.Value)
			case GTE: // greater than or equal
				whereCondition = where.Key + ":[" + fmt.Sprintf("%v", where.Value) + " TO *]"
			//whereCondition = r.Row.Field(where.Key).Ge(where.Value)
			// case IN: // TODO: implement!!!!
			default:
				return query, errors.New(fmt.Sprintf("invalid comparison operator: %v", where.RelationalOperator))
			}

			if index == 0 {
				whereStmt = whereCondition
				//if where.LogicalOperator == NOT {
				//whereStmt = whereStmt.Not()
				//}
			} else {
				switch where.LogicalOperator {
				case AND:
					whereStmt = whereStmt + " AND " + whereCondition
					//whereStmt = whereStmt.And(whereCondition)
				case OR:
					whereStmt = whereStmt + " OR " + whereCondition
				//whereStmt = whereStmt.Or(whereCondition)
				////case NOT:
				////whereStmt = whereStmt.And(whereCondition).Not()
				default:
					whereStmt = whereStmt + " AND " + whereCondition
					//whereStmt = whereStmt.And(whereCondition)
				}
			}
		}

		// TODO: delete!!
		log.Printf("DbSearch whereStmt: %s", whereStmt)
		//query = query.Filter(whereStmt)
		//query = query.Filter(whereStmt)
	}

	return whereStmt, nil
}

//func processAggregations(query r.Term, ar *ArRethinkDb) (r.Term, error) {
//// sum
//if sum := ar.Query().Aggregations[SUM]; sum != nil {
//if len(sum) == 1 {
//query = query.Sum(sum...)
//} else {
//return query, errors.New(fmt.Sprintf("rethinkdb does not support summing more than one field at a time: %v", sum))
//}
//}

//// distinct
//if ar.Query().Distinct {
//query = query.Distinct()
//}

//return query, nil
//}

func processSorts(ar *ArOrchestrate) (sort string) {
	if len(ar.Query().OrderBys) > 0 {
		sort = ""

		for i, orderBy := range ar.Query().OrderBys {
			if i > 0 {
				sort += ","
			}

			sort += "value." + orderBy.Key + ":"

			switch orderBy.SortOrder {
			case DESC: // descending
				sort += "desc"
			default: // ascending
				sort += "asc"
			}
		}
	}

	return sort
}
