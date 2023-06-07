package Helpers

import (
	"Weaviate/Constants"
	"sort"
)

/*
CalculateL2Norm
Calculates the Euclidean distance between two vectors. In this version, the
square root is omitted, increasing the performance even more.
*/
func CalculateL2Norm(vector1, vector2 *[]int8, size uint16) float64 {
	sumOfSquares := 0.0
	var i uint16 = 0
	for i = 0; i < size; i++ {
		diff := float64((*vector1)[i] - (*vector2)[i])
		sumOfSquares += diff * diff
	}
	return sumOfSquares
}

/*
SortBasedOn
Sorts an array "points" in ascending order, based on the corresponding values in "keys"
For example:
before:

	points = {1, 2, 3, 4}
	keys = {5, 2, 4, 3}

after:

	points = {2, 4, 3, 1}
	keys = {2, 3, 4, 5}
*/
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
