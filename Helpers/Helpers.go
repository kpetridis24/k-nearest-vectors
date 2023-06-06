package Helpers

import (
	"Weaviate/Constants"
	"sort"
)

func CalculateL2Norm(vector1, vector2 *[]int8, size uint16) float64 {
	sumOfSquares := 0.0
	var i uint16 = 0
	for i = 0; i < size; i++ {
		diff := float64((*vector1)[i] - (*vector2)[i])
		sumOfSquares += diff * diff
	}
	return sumOfSquares
}

func SortBasedOn(points *[]int8, keys *[]float64) {
	numPoints := len(*points) / int(Constants.NumOfDimensions)
	indices := make([]int, numPoints)
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return (*keys)[indices[i]] < (*keys)[indices[j]]
	})

	sortedPoints := make([]int8, 0, len(*points))
	sortedKeys := make([]float64, numPoints)

	for index, sortedIndex := range indices {
		sortedKeys[index] = (*keys)[sortedIndex]
		temp := (*points)[sortedIndex*int(Constants.NumOfDimensions) : (sortedIndex+1)*int(Constants.NumOfDimensions)]
		sortedPoints = append(sortedPoints, temp...)
	}

	copy(*keys, sortedKeys)
	copy(*points, sortedPoints)
}
