package Helpers

import (
	"Weaviate/Constants"
	"fmt"
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

func ValidateKnnResults(vptResult *MaxPriorityQueue, parallelResult, naiveResult []float64) {
	vptResultAsArray := make([]float64, Constants.K)
	for i := 0; i < Constants.K; i++ {
		d := vptResult.Pop().DistanceFromQuery
		vptResultAsArray[i] = d
	}

	for i := 0; i < Constants.K; i++ {
		if vptResultAsArray[Constants.K-1-i] != parallelResult[i] ||
			vptResultAsArray[Constants.K-1-i] != naiveResult[i] ||
			parallelResult[i] != naiveResult[i] {

			fmt.Println(vptResultAsArray)
			fmt.Println(parallelResult)
			fmt.Println(naiveResult)
			panic("[ERROR]: K nearest neighbors not consistent amongst methods")
		}
	}
}

type KnnQueueItem struct {
	KnnVector         *[]int8
	DistanceFromQuery float64
}

// MaxPriorityQueue using binary max heap maintain structure
type MaxPriorityQueue struct {
	array              []KnnQueueItem
	nextAvailableIndex int
	capacity           int
}

// NewMaxPriorityQueue constructor to initialise a priority queue of capacity maximumCapacity
func NewMaxPriorityQueue(maximumCapacity int) *MaxPriorityQueue {
	return &MaxPriorityQueue{
		array:              make([]KnnQueueItem, maximumCapacity),
		nextAvailableIndex: 0,
		capacity:           maximumCapacity,
	}
}

func (queue *MaxPriorityQueue) Insert(newItem *KnnQueueItem) {
	queue.array[queue.nextAvailableIndex] = *newItem
	queue.nextAvailableIndex++
	currentIndex := queue.nextAvailableIndex - 1
	parentIndex := (currentIndex - 1) / 2

	for parentIndex > 0 &&
		queue.array[currentIndex].DistanceFromQuery > queue.array[parentIndex].DistanceFromQuery {
		queue.array[parentIndex], queue.array[currentIndex] = queue.array[currentIndex], queue.array[parentIndex]
		currentIndex = parentIndex
		parentIndex = (currentIndex - 1) / 2
	}

	// if queue exceeds maximum capacity, pop the max item
	if queue.nextAvailableIndex == *queue.Capacity() {
		queue.Pop()
	}
}

func (queue *MaxPriorityQueue) Pop() *KnnQueueItem {
	if queue.nextAvailableIndex == 1 {
		queue.nextAvailableIndex = 0
		return &queue.array[0]
	}

	var root = queue.array[0]
	queue.array[0] = queue.array[queue.nextAvailableIndex-1]
	queue.nextAvailableIndex--

	// heapify
	currentIndex := 0
	leftChildIndex := 2*currentIndex + 1
	rightChildIndex := 2*currentIndex + 2
	maxItemIndex := currentIndex

	for true {
		if leftChildIndex <= queue.Len() &&
			queue.array[leftChildIndex].DistanceFromQuery > queue.array[maxItemIndex].DistanceFromQuery {
			maxItemIndex = leftChildIndex
		}
		if rightChildIndex <= queue.Len() &&
			queue.array[rightChildIndex].DistanceFromQuery > queue.array[maxItemIndex].DistanceFromQuery {
			maxItemIndex = rightChildIndex
		}

		if maxItemIndex == currentIndex {
			return &root
		}

		queue.array[currentIndex], queue.array[maxItemIndex] = queue.array[maxItemIndex], queue.array[currentIndex]

		currentIndex = maxItemIndex
		leftChildIndex = 2*currentIndex + 1
		rightChildIndex = 2*currentIndex + 2
		maxItemIndex = currentIndex
	}

	return &root
}

func (queue *MaxPriorityQueue) Peak() *KnnQueueItem {
	return &queue.array[0]
}

func (queue *MaxPriorityQueue) Len() int {
	return queue.nextAvailableIndex
}

func (queue *MaxPriorityQueue) Capacity() *int {
	return &queue.capacity
}
