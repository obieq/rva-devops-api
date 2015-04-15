package rethinkdb

import (
	"runtime"

	. "github.com/obieq/goar"
	. "github.com/obieq/goar/tests/models"

	r "github.com/dancannon/gorethink"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type RethinkDbAutomobile struct {
	ArRethinkDb
	Automobile
	Id           string `gorethink:"id,omitempty"`
	SafetyRating int
	//Timestamps
}

// Model for testing RethinkDb ActiveRecord error conditions
type ErrorTestingModel struct {
	ArRethinkDb
}

func (model RethinkDbAutomobile) ToActiveRecord() *RethinkDbAutomobile {
	return ToAR(&model).(*RethinkDbAutomobile)
}

func (model ErrorTestingModel) ToActiveRecord() *ErrorTestingModel {
	return ToAR(&model).(*ErrorTestingModel)
}

func (m *RethinkDbAutomobile) Validate() {
	m.Validation.Required("Year", m.Year)
	m.Validation.Required("Make", m.Make)
	m.Validation.Required("Model", m.Model)
}

func (m *ErrorTestingModel) Validate() {}

var _ = Describe("ActiveRecord", func() {

	var (
		DbModel *RethinkDbAutomobile = RethinkDbAutomobile{}.ToActiveRecord()
		ModelS  *RethinkDbAutomobile
		MK      *RethinkDbAutomobile
		Sprite  *RethinkDbAutomobile
	)

	var errorConnectOpts = func() r.ConnectOpts {
		return r.ConnectOpts{
			Address: "",
		}
	}

	Context("Validations", func() {
		It("should require properties", func() {
			ModelS = RethinkDbAutomobile{}.ToActiveRecord()
			ModelS.Make = "tesla"
			Ω(ModelS.Valid()).Should(BeFalse())
		})

		// TODO: move to active_record_test.go
		It("should set self", func() {
			ModelS = RethinkDbAutomobile{}.ToActiveRecord()
			ModelS.Make = "tesla"
			Ω("tesla").Should(Equal(ModelS.Self().(*RethinkDbAutomobile).Make))

			ModelS.Make = "porsche"
			Ω("porsche").Should(Equal(ModelS.Self().(*RethinkDbAutomobile).Make))
		})
	})

	Context("Error Handling", func() {
		It("should return an error when the Truncate() method is called", func() {
			//truncate = func(modelName string) (*r.ResultRows, error) {
			//return nil, errors.New("some error")
			//}
			//_, err := Model.Truncate()
			errorModel := ErrorTestingModel{}.ToActiveRecord()
			_, err := errorModel.Truncate()
			Ω(err).ShouldNot(BeNil())
		})

		It("should return an error when the Find() method is called", func() {
			errorModel := ErrorTestingModel{}.ToActiveRecord()
			_, err := errorModel.Find("id_does_not_exist")
			Ω(err).ShouldNot(BeNil())
		})

		It("should return an error when the All() method is called", func() {
			var results []ErrorTestingModel
			errorModel := ErrorTestingModel{}.ToActiveRecord()
			err := errorModel.All(results, nil)
			Ω(err).ShouldNot(BeNil())
		})

		It("should return an error when an invalid logical operator is used", func() {
			var results []ErrorTestingModel
			errorModel := ErrorTestingModel{}.ToActiveRecord()
			err := errorModel.Where(QueryCondition{Key: "Doesn't Matter", RelationalOperator: EQ + 5000, Value: "Doesn't Matter"}).Run(&results)
			Ω(err).ShouldNot(BeNil())
		})

		It("should return an error when the DbSearch() method is called", func() {
			var results []ErrorTestingModel
			errorModel := ErrorTestingModel{}.ToActiveRecord()
			err := errorModel.Where(QueryCondition{Key: "Doesn't Matter", RelationalOperator: EQ, Value: "Doesn't Matter"}).Run(&results)
			Ω(err).ShouldNot(BeNil())
		})

		It("should return an error when calling the Connect method", func() {
			var err error
			defer func() {
				if e := recover(); e != nil {
					if _, ok := e.(runtime.Error); ok {
						panic(e)
					}
					err = e.(error)
				}
			}()
			connectOpts = errorConnectOpts
			connect()
			Ω(err).To(HaveOccurred())
		})
	})
	Context("DB Interactions", func() {
		BeforeEach(func() {
			DbModel.Truncate() // delete all records created during previous test

			ModelS = RethinkDbAutomobile{Id: "id1", SafetyRating: 5, Automobile: Automobile{Vehicle: Vehicle{Make: "tesla", Year: 2014, Model: "model s"}}}.ToActiveRecord()
			Ω(ModelS.Valid()).Should(BeTrue())

			MK = RethinkDbAutomobile{Id: "id2", SafetyRating: 3, Automobile: Automobile{Vehicle: Vehicle{Make: "austin healey", Year: 1960, Model: "3000"}}}.ToActiveRecord()
			Ω(MK.Valid()).Should(BeTrue())

			Sprite = RethinkDbAutomobile{Id: "id3", SafetyRating: 2, Automobile: Automobile{Vehicle: Vehicle{Make: "austin healey", Year: 1960, Model: "sprite"}}}.ToActiveRecord()
			Ω(ModelS.Valid()).Should(BeTrue())
		})

		Context("Persistance", func() {
			It("should persist a new model and find it by id", func() {
				Ω(ModelS.Save()).Should(BeTrue())

				result, _ := RethinkDbAutomobile{}.ToActiveRecord().Find(ModelS.Id)
				Ω(result).ShouldNot(BeNil())
				model := result.(*RethinkDbAutomobile)
				Ω(model.Id).Should(Equal(ModelS.Id))
			})

			It("should update an existing model", func() {
				Ω(ModelS.Save()).Should(BeTrue())
				year := ModelS.Year
				modelName := ModelS.Model

				// create
				result, _ := RethinkDbAutomobile{}.ToActiveRecord().Find(ModelS.Id)
				Ω(result).ShouldNot(BeNil())
				dbModel := result.(*RethinkDbAutomobile).ToActiveRecord()
				Ω(dbModel.Id).Should(Equal(ModelS.Id))
				Ω(dbModel.CreatedAt).ShouldNot(BeNil())
				Ω(dbModel.UpdatedAt).Should(BeNil())

				// update
				dbModel.Year += 1
				dbModel.Model += " updated"
				Ω(dbModel.Save()).Should(BeTrue())

				// verify updates
				result, _ = RethinkDbAutomobile{}.ToActiveRecord().Find(ModelS.Id)
				Ω(result).ShouldNot(BeNil())
				Ω(dbModel.Year).Should(Equal(year + 1))
				Ω(dbModel.Model).Should(Equal(modelName + " updated"))
				Ω(dbModel.CreatedAt).ShouldNot(BeNil())
				Ω(dbModel.UpdatedAt).ShouldNot(BeNil())
			})

			It("should delete an existing model", func() {
				// create and verify
				Ω(ModelS.Save()).Should(BeTrue())
				result, _ := RethinkDbAutomobile{}.ToActiveRecord().Find(ModelS.Id)
				Ω(result).ShouldNot(BeNil())

				// delete
				err := ModelS.Delete()
				Ω(err).NotTo(HaveOccurred())

				// verify delete
				result, _ = RethinkDbAutomobile{}.ToActiveRecord().Find(ModelS.Id)
				Ω(result).Should(BeNil())
			})
		})

		Context("Querying", func() {
			It("should return all models for a given type", func() {
				Ω(ModelS.Save()).Should(BeTrue())
				Ω(MK.Save()).Should(BeTrue())

				var results []RethinkDbAutomobile
				DbModel.All(&results, nil)
				Ω(len(results)).Should(Equal(2))

				ids := []string{ModelS.Id, MK.Id}
				for _, model := range results {
					//model := DbModel.(*RethinkDbAutomobile)
					Ω(ids).Should(ContainElement(model.Id))
				}
			})

			Context("Relational Operators", func() {
				Context("Equal", func() {
					It("should query with two EQ operators", func() {
						var results []RethinkDbAutomobile
						Ω(MK.Save()).Should(BeTrue())
						Ω(Sprite.Save()).Should(BeTrue())

						ar := RethinkDbAutomobile{}.ToActiveRecord()
						ar.Where(QueryCondition{Key: "Year", RelationalOperator: EQ, Value: 1960})
						err := ar.Where(QueryCondition{Key: "Model", RelationalOperator: EQ, Value: "sprite"}).Run(&results)

						Ω(err).NotTo(HaveOccurred())
						Ω(results).ShouldNot(BeNil())
						Ω(len(results)).Should(Equal(1))

						auto := results[0]
						Ω(auto).ShouldNot(BeNil())
						Ω(auto.Model).Should(Equal("sprite"))
					})
				})

				Context("Not Equal", func() {
					It("should query with two NE operators", func() {
						var results []RethinkDbAutomobile
						Ω(ModelS.Save()).Should(BeTrue())
						Ω(MK.Save()).Should(BeTrue())
						Ω(Sprite.Save()).Should(BeTrue())

						ar := RethinkDbAutomobile{}.ToActiveRecord()
						ar.Where(QueryCondition{Key: "Year", RelationalOperator: NE, Value: 2014})
						err := ar.Where(QueryCondition{Key: "Model", RelationalOperator: NE, Value: "sprite"}).Run(&results)

						Ω(err).NotTo(HaveOccurred())
						Ω(results).ShouldNot(BeNil())
						Ω(len(results)).Should(Equal(1))

						auto := results[0]
						Ω(auto).ShouldNot(BeNil())
						Ω(auto.Model).Should(Equal("3000"))
					})
				})

				Context("Greater Than", func() {
					It("should query with two GT operators", func() {
						var results []RethinkDbAutomobile
						Ω(ModelS.Save()).Should(BeTrue())
						Ω(MK.Save()).Should(BeTrue())
						Ω(Sprite.Save()).Should(BeTrue())

						ar := RethinkDbAutomobile{}.ToActiveRecord()
						ar.Where(QueryCondition{Key: "Year", RelationalOperator: GT, Value: 1960})
						err := ar.Where(QueryCondition{Key: "Make", RelationalOperator: GT, Value: "porsche"}).Run(&results)

						Ω(err).NotTo(HaveOccurred())
						Ω(results).ShouldNot(BeNil())
						Ω(len(results)).Should(Equal(1))

						auto := results[0]
						Ω(auto).ShouldNot(BeNil())
						Ω(auto.Make).Should(Equal("tesla"))
					})
				})

				Context("Greater Than Equal", func() {
					It("should query with two GTE operators", func() {
						var results []RethinkDbAutomobile
						Ω(ModelS.Save()).Should(BeTrue())
						Ω(MK.Save()).Should(BeTrue())
						Ω(Sprite.Save()).Should(BeTrue())

						ar := RethinkDbAutomobile{}.ToActiveRecord()
						ar.Where(QueryCondition{Key: "Year", RelationalOperator: GTE, Value: 1960})
						err := ar.Where(QueryCondition{Key: "Make", RelationalOperator: GTE, Value: "austin healey"}).Run(&results)

						Ω(err).NotTo(HaveOccurred())
						Ω(results).ShouldNot(BeNil())
						Ω(len(results)).Should(Equal(3))
					})
				})

				Context("Less Than", func() {
					It("should query with two LT operators", func() {
						var results []RethinkDbAutomobile
						Ω(ModelS.Save()).Should(BeTrue())
						Ω(MK.Save()).Should(BeTrue())
						Ω(Sprite.Save()).Should(BeTrue())

						ar := RethinkDbAutomobile{}.ToActiveRecord()
						ar.Where(QueryCondition{Key: "Year", RelationalOperator: LT, Value: 1961})
						err := ar.Where(QueryCondition{Key: "Model", RelationalOperator: LT, Value: "sprite"}).Run(&results)

						Ω(err).NotTo(HaveOccurred())
						Ω(results).ShouldNot(BeNil())
						Ω(len(results)).Should(Equal(1))

						auto := results[0]
						Ω(auto).ShouldNot(BeNil())
						Ω(auto.Model).Should(Equal("3000"))
					})
				})

				Context("Less Than Equal", func() {
					It("should query with two LT operators", func() {
						var results []RethinkDbAutomobile
						Ω(ModelS.Save()).Should(BeTrue())
						Ω(MK.Save()).Should(BeTrue())
						Ω(Sprite.Save()).Should(BeTrue())

						ar := RethinkDbAutomobile{}.ToActiveRecord()
						ar.Where(QueryCondition{Key: "Year", RelationalOperator: LTE, Value: 1960})
						err := ar.Where(QueryCondition{Key: "Model", RelationalOperator: LTE, Value: "sprite"}).Run(&results)

						Ω(err).NotTo(HaveOccurred())
						Ω(results).ShouldNot(BeNil())
						Ω(len(results)).Should(Equal(2))
					})
				})
			})
		})

		Context("Logical Operators", func() {
			Context("And", func() {
				It("should query with two AND operators", func() {
					Ω(ModelS.Save()).Should(BeTrue()) // year => 1960
					Ω(MK.Save()).Should(BeTrue())     // year => 1960
					Ω(Sprite.Save()).Should(BeTrue()) // year => 1960

					ar := RethinkDbAutomobile{}.ToActiveRecord()
					var results []RethinkDbAutomobile
					ar.Where(QueryCondition{Key: "Year", RelationalOperator: EQ, Value: 1960})
					err := ar.Where(QueryCondition{LogicalOperator: AND, Key: "Model", RelationalOperator: EQ, Value: "sprite"}).Run(&results)

					Ω(err).NotTo(HaveOccurred())
					Ω(results).ShouldNot(BeNil())
					Ω(len(results)).Should(Equal(1))

					auto := results[0]
					Ω(auto).ShouldNot(BeNil())
					Ω(auto.Model).Should(Equal("sprite"))
				})
			})

			Context("Or", func() {
				It("should query with two OR operators", func() {
					Ω(MK.Save()).Should(BeTrue())     // year => 1960
					Ω(Sprite.Save()).Should(BeTrue()) // year => 1960

					ar := RethinkDbAutomobile{}.ToActiveRecord()
					var results []RethinkDbAutomobile
					ar.Where(QueryCondition{Key: "Year", RelationalOperator: EQ, Value: 1960})
					ar.Where(QueryCondition{LogicalOperator: OR, Key: "Year", RelationalOperator: EQ, Value: "3000"})
					err := ar.Where(QueryCondition{LogicalOperator: OR, Key: "Model", RelationalOperator: EQ, Value: "invalid model name"}).Run(&results)

					Ω(err).NotTo(HaveOccurred())
					Ω(results).ShouldNot(BeNil())
					Ω(len(results)).Should(Equal(2))
				})
			})

			//Context("Not", func() {
			//It("should query with two NOT operators", func() {
			//Ω(ModelS.Save()).Should(BeTrue()) // year => 1960
			//Ω(MK.Save()).Should(BeTrue())     // year => 1960
			//Ω(Sprite.Save()).Should(BeTrue()) // year => 1960

			//ar := RethinkDbAutomobile{}.ToActiveRecord()
			//var results []RethinkDbAutomobile
			//ar.Where(QueryCondition{LogicalOperator: NOT, Key: "Year", RelationalOperator: EQ, Value: 2014})
			//err := ar.Where(QueryCondition{LogicalOperator: NOT, Key: "Model", RelationalOperator: EQ, Value: "3000"}).Run(&results)

			//Ω(err).NotTo(HaveOccurred())
			//Ω(results).ShouldNot(BeNil())
			//Ω(len(results)).Should(Equal(1))

			//auto := results[0]
			//Ω(auto).ShouldNot(BeNil())
			//Ω(auto.Model).Should(Equal("sprite"))
			//})
			//})
		})

		Context("Query Transformations", func() {
			Context("Order Bys", func() {
				It("should order DESC", func() {
					Ω(ModelS.Save()).Should(BeTrue()) // year => 1960
					Ω(MK.Save()).Should(BeTrue())     // year => 1960
					Ω(Sprite.Save()).Should(BeTrue())

					ar := RethinkDbAutomobile{}.ToActiveRecord()
					var results []RethinkDbAutomobile
					ar.Order(OrderBy{Key: "Year", SortOrder: DESC})
					err := ar.Order(OrderBy{Key: "Model", SortOrder: ASC}).Run(&results)

					Ω(err).NotTo(HaveOccurred())
					Ω(results).ShouldNot(BeNil())
					Ω(len(results)).Should(Equal(3))

					Ω(results[0].Model).Should(Equal("model s"))
					Ω(results[1].Model).Should(Equal("3000"))
					Ω(results[2].Model).Should(Equal("sprite"))
				})
			})

			Context("Plucks", func() {
				It("should pluck a single field", func() {
					Ω(ModelS.Save()).Should(BeTrue()) // year => 1960
					Ω(MK.Save()).Should(BeTrue())     // year => 1960
					Ω(Sprite.Save()).Should(BeTrue())

					ar := RethinkDbAutomobile{}.ToActiveRecord()
					var results []RethinkDbAutomobile
					err := ar.Pluck("Year").Run(&results)

					Ω(err).NotTo(HaveOccurred())
					Ω(results).ShouldNot(BeNil())
					Ω(len(results)).Should(Equal(3))

					Ω(results[0].Year).ShouldNot(BeNil())
					Ω(results[0].Model).Should(Equal(""))
					Ω(results[1].Year).ShouldNot(BeNil())
					Ω(results[1].Model).Should(Equal(""))
					Ω(results[2].Year).ShouldNot(BeNil())
					Ω(results[2].Model).Should(Equal(""))
				})

				It("should pluck multiple fields", func() {
					It("should pluck multiple fields", func() {
						Ω(ModelS.Save()).Should(BeTrue()) // year => 1960
						Ω(MK.Save()).Should(BeTrue())     // year => 1960
						Ω(Sprite.Save()).Should(BeTrue())

						ar := RethinkDbAutomobile{}.ToActiveRecord()
						var results []RethinkDbAutomobile
						err := ar.Pluck("Year", "Model").Run(&results)

						Ω(err).NotTo(HaveOccurred())
						Ω(results).ShouldNot(BeNil())
						Ω(len(results)).Should(Equal(3))

						Ω(results[0].Year).ShouldNot(BeNil())
						Ω(results[0].Model).ShouldNot(BeNil())
						Ω(results[1].Year).ShouldNot(BeNil())
						Ω(results[1].Model).ShouldNot(BeNil())
						Ω(results[2].Year).ShouldNot(BeNil())
						Ω(results[2].Model).ShouldNot(BeNil())
					})
				})
			})

			Context("Aggregations", func() {
				It("should SUM a single field", func() {
					Ω(ModelS.Save()).Should(BeTrue()) // year => 2014
					Ω(MK.Save()).Should(BeTrue())     // year => 1960
					Ω(Sprite.Save()).Should(BeTrue()) // year => 1960

					ar := RethinkDbAutomobile{}.ToActiveRecord()
					var results []interface{}
					err := ar.Sum("Year").Run(&results)

					Ω(err).NotTo(HaveOccurred())
					Ω(results).ShouldNot(BeNil())
					Ω(len(results)).Should(Equal(1))
					Ω(int(results[0].(float64))).Should(Equal(ModelS.Year + MK.Year + Sprite.Year))
				})

				It("should perform a DISTINCT query", func() {
					Ω(MK.Save()).Should(BeTrue())     // year => 1960
					Ω(Sprite.Save()).Should(BeTrue()) // year => 1960

					ar := RethinkDbAutomobile{}.ToActiveRecord()
					var results []RethinkDbAutomobile
					err := ar.Pluck("Year").Distinct().Run(&results)

					Ω(err).NotTo(HaveOccurred())
					Ω(results).ShouldNot(BeNil())
					Ω(len(results)).Should(Equal(1))
					Ω(results[0].Year).Should(Equal(MK.Year))
					Ω(results[0].Model).Should(Equal(""))
				})
			})

			// rethinkdb only allows one field or function to be summed for a given query
			It("should not SUM multiple fields", func() {
				Ω(ModelS.Save()).Should(BeTrue()) // year => 1960
				Ω(MK.Save()).Should(BeTrue())     // year => 1960
				Ω(Sprite.Save()).Should(BeTrue())

				ar := RethinkDbAutomobile{}.ToActiveRecord()
				var results []interface{}
				err := ar.Sum("Year", "SafetyRating").Run(&results)

				Ω(err).To(HaveOccurred())
			})
		})
	})
})
