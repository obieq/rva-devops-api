package dynamodb

import (
	"errors"
	"runtime"

	. "github.com/obieq/goar"
	. "github.com/obieq/goar/tests/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dynamodb", func() {
	var (
		ModelS, MK, Sprite, Panamera, Evoque, Bugatti DynamodbAutomobile
		ar                                            *DynamodbAutomobile
	)

	var errEnvVars = func() (map[string]string, error) {
		return nil, errors.New("EnvVars Error")
	}

	BeforeEach(func() {
		ar = DynamodbAutomobile{}.ToActiveRecord()
	})

	It("should initialize client", func() {
		client := Client()
		Ω(client).ShouldNot(BeNil())
	})

	Context("DB Interactions", func() {
		BeforeEach(func() {
			ModelS = DynamodbAutomobile{SafetyRating: 5, Automobile: Automobile{Vehicle: Vehicle{Make: "tesla", Year: 2009, Model: "model s"}}}
			ToAR(&ModelS)
			ModelS.SetKey("id1")
			Ω(ModelS.Valid()).Should(BeTrue())

			MK = DynamodbAutomobile{SafetyRating: 3, Automobile: Automobile{Vehicle: Vehicle{Make: "austin healey", Year: 1960, Model: "3000"}}}
			ToAR(&MK)
			MK.SetKey("id2")
			Ω(MK.Valid()).Should(BeTrue())

			Sprite = DynamodbAutomobile{SafetyRating: 2, Automobile: Automobile{Vehicle: Vehicle{Make: "austin healey", Year: 1960, Model: "sprite"}}}
			ToAR(&Sprite)
			Sprite.SetKey("id3")
			Ω(Sprite.Valid()).Should(BeTrue())

			Panamera = DynamodbAutomobile{SafetyRating: 5, Automobile: Automobile{Vehicle: Vehicle{Make: "porsche", Year: 2010, Model: "panamera"}}}
			ToAR(&Panamera)
			Panamera.SetKey("id4")
			Ω(Panamera.Valid()).Should(BeTrue())

			Evoque = DynamodbAutomobile{SafetyRating: 1, Automobile: Automobile{Vehicle: Vehicle{Make: "land rover", Year: 2013, Model: "evoque"}}}
			ToAR(&Evoque)
			Evoque.SetKey("id5")
			Ω(Evoque.Valid()).Should(BeTrue())

			Bugatti = DynamodbAutomobile{SafetyRating: 4, Automobile: Automobile{Vehicle: Vehicle{Make: "bugatti", Year: 2013, Model: "veyron"}}}
			ToAR(&Bugatti)
			Bugatti.SetKey("id6")
			Ω(Bugatti.Valid()).Should(BeTrue())
		})

		Context("Error Handling", func() {
			It("should return an error when calling envVars", func() {
				// swallow the panic method that occurs during this test
				var err error
				defer func() {
					if e := recover(); e != nil {
						if _, ok := e.(runtime.Error); ok {
							panic(e)
						}
						err = e.(error)
					}
				}()
				envVars = errEnvVars
				connect()
				Ω(err).To(HaveOccurred())
			})

			It("should return an error when the Truncate() method is called", func() {
				auto := DynamodbAutomobile{}.ToActiveRecord()
				_, err := auto.Truncate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should return an error when the All() method is called", func() {
				auto := DynamodbAutomobile{}.ToActiveRecord()
				err := auto.All(auto, nil)
				Ω(err).ShouldNot(BeNil())
			})

			It("should return an error when the Search() method is called", func() {
				auto := DynamodbAutomobile{}.ToActiveRecord()
				err := auto.DbSearch(auto)
				Ω(err).ShouldNot(BeNil())
			})

			It("should return an error when trying to find an ID that doesn't exist", func() {
				auto, err := DynamodbAutomobile{}.ToActiveRecord().Find("does not exist")
				Expect(err).To(HaveOccurred())
				Ω(auto).Should(BeNil())
			})

			It("should return an error when trying to patch an ID that doesn't exist", func() {
				auto := DynamodbAutomobile{}.ToActiveRecord()
				auto.SetKey("does not exist")
				success, err := auto.Patch()
				Expect(err).To(HaveOccurred())
				Ω(success).Should(BeFalse())
			})
		})

		Context("Persistance", func() {
			It("should create a model and find it by id", func() {
				success, err := ModelS.Save()

				Ω(ModelS.ModelName()).Should(Equal("DynamodbAutomobiles"))
				Ω(err).Should(BeNil())
				Ω(success).Should(BeTrue())

				result, err := DynamodbAutomobile{}.ToActiveRecord().Find(ModelS.ID)
				Ω(err).Should(BeNil())
				Ω(result).ShouldNot(BeNil())
				model := result.(*DynamodbAutomobile)
				Ω(model.ID).Should(Equal(ModelS.ID))
				Ω(model.CreatedAt).ShouldNot(BeNil())
			})

			It("should update an existing model", func() {
				Sprite.Delete()
				Ω(Sprite.Save()).Should(BeTrue())
				year := Sprite.Year
				modelName := Sprite.Model

				// create
				result, _ := ar.Find(Sprite.ID)
				Ω(result).ShouldNot(BeNil())
				dbModel := result.(*DynamodbAutomobile).ToActiveRecord()
				Ω(dbModel.ID).Should(Equal(Sprite.ID))
				Ω(dbModel.CreatedAt).ShouldNot(BeNil())
				Ω(dbModel.UpdatedAt).Should(BeNil())

				// update
				dbModel.Year += 1
				dbModel.Model += " updated"

				success, err := dbModel.Save()
				Ω(err).Should(BeNil())
				Ω(success).Should(BeTrue())

				// verify updates
				result, err = ar.Find(Sprite.ID)
				Expect(err).NotTo(HaveOccurred())
				Ω(result).ShouldNot(BeNil())
				Ω(dbModel.Year).Should(Equal(year + 1))
				Ω(dbModel.Model).Should(Equal(modelName + " updated"))
				Ω(dbModel.CreatedAt).ShouldNot(BeNil())
				Ω(dbModel.UpdatedAt).ShouldNot(BeNil())
			})

			It("should perform partial (patch) updates", func() {
				Sprite.Delete()

				// create
				Ω(Sprite.Save()).Should(BeTrue())
				year := Sprite.Year
				modelName := Sprite.Model
				safetyRating := Sprite.SafetyRating

				// verify
				result, _ := ar.Find(Sprite.ID)
				Ω(result).ShouldNot(BeNil())
				dbModel := result.(*DynamodbAutomobile).ToActiveRecord()
				Ω(dbModel.ID).Should(Equal(Sprite.ID))
				Ω(dbModel.CreatedAt).ShouldNot(BeNil())
				Ω(dbModel.UpdatedAt).Should(BeNil())

				// partial update
				s2 := DynamodbAutomobile{SafetyRating: safetyRating + 1}.ToActiveRecord()
				s2.SetKey(Sprite.ID)
				success, err := s2.Patch()
				Ω(err).Should(BeNil())
				Ω(s2.Validation.Errors).Should(BeNil())
				Ω(success).Should(BeTrue())

				// verify updates
				result, err = ar.Find(Sprite.ID)
				Expect(err).NotTo(HaveOccurred())
				Ω(result).ShouldNot(BeNil())
				dbModel = result.(*DynamodbAutomobile).ToActiveRecord()
				Ω(dbModel.Year).Should(Equal(year))                     // should be no change
				Ω(dbModel.Model).Should(Equal(modelName))               // should be no change
				Ω(dbModel.SafetyRating).Should(Equal(safetyRating + 1)) // should have incremented by one
				Ω(dbModel.CreatedAt).ShouldNot(BeNil())                 // should be no change
				Ω(dbModel.UpdatedAt).ShouldNot(BeNil())                 // should have been set via active record framework
			})

			It("should delete an existing model", func() {
				// create and verify
				Ω(MK.Save()).Should(BeTrue())
				result, err := ar.Find(MK.ID)
				Expect(err).NotTo(HaveOccurred())
				Ω(result).ShouldNot(BeNil())
				Ω(MK.ID).Should(Equal(result.(*DynamodbAutomobile).ID))

				// delete
				err = MK.Delete()
				Ω(err).NotTo(HaveOccurred())

				// verify delete
				result, _ = ar.Find(MK.ID)
				Ω(result).Should(BeNil())
			})
		})
	})
})
