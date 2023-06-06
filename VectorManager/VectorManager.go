package VectorManager

import (
	"Weaviate/Constants"
	"math/rand"
	"time"
)

const (
	NumOfVectors    = Constants.NumOfVectors
	NumOfDimensions = Constants.NumOfDimensions
	Min             = Constants.Min
	Max             = Constants.Max
)

type RowMajorVectorManager interface {
	FillVector(vector *[]int8)
	GenerateVectors() *[]int8
	GenerateSingleVector() *[]int8
}

type TwoDimVectorManager interface {
	Transfer2D() *[][]int8
}

type Random2DVectorManager struct{}
type RandomRowMajorVectorManager struct{}

func (generator Random2DVectorManager) Transfer2D(rowMajorVectors *[]int8) *[][]int8 {
	vectors := make([][]int8, NumOfVectors)
	for i := range vectors {
		vectors[i] = make([]int8, NumOfDimensions)
		copy(vectors[i], (*rowMajorVectors)[i*int(NumOfDimensions):(i+1)*int(NumOfDimensions)])
	}
	return &vectors
}

func (generator RandomRowMajorVectorManager) GenerateVectors() *[]int8 {
	vectors := make([]int8, NumOfVectors*uint32(NumOfDimensions))
	generator.FillVector(&vectors)
	return &vectors
}

func (generator RandomRowMajorVectorManager) GenerateSingleVector() *[]int8 {
	vector := make([]int8, uint32(NumOfDimensions))
	generator.FillVector(&vector)
	return &vector
}

func (generator RandomRowMajorVectorManager) FillVector(vector *[]int8) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len(*vector); i++ {
		(*vector)[i] = int8(rand.Intn(Max-Min+1)) + Min
	}
}
