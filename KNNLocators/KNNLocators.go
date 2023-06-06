package KNNLocators

import (
	"Weaviate/Constants"
	util "Weaviate/Helpers"
	"math"
	"sort"
	"sync"
)

const (
	NumOfVectors    = Constants.NumOfVectors
	NumOfDimensions = Constants.NumOfDimensions
	MaxCPUs         = Constants.MaxCPUs
)

type VPTreeNode struct {
	vantagePoint *[]int8
	radius       float64
	inside       *VPTreeNode
	outside      *VPTreeNode
}

type KnnLocator interface {
	BuildIndex(points *[]int8, numPoints int) *VPTreeNode
	SearchKNearest(points *[]int8, query *[]int8, k int) []float64
	SearchKNearestNaive(points *[][]int8, query *[]int8, k int) []float64
}

type NaiveKnnExposer interface {
	KnnLocator
	SearchKNearestNaive(points *[][]int8, query *[]int8, k int) []float64
}

type NaiveKnnLocator struct{}

type ParallelKnnLocator struct{}

type VPTreeKnnLocator struct{}

func (locator NaiveKnnLocator) SearchKNearestNaive(points *[][]int8, query *[]int8, k int) []float64 {
	distances := make([]float64, NumOfVectors)
	for index, vector := range *points {
		distances[index] = util.CalculateL2Norm(query, &vector, NumOfDimensions)
	}
	sort.Float64s(distances[:])
	return distances[:k]
}

func (locator ParallelKnnLocator) SearchKNearest(points, query *[]int8, k int) []float64 {
	distances := make([]float64, NumOfVectors)
	wg := sync.WaitGroup{}
	wg.Add(MaxCPUs)

	for i := 0; i < MaxCPUs; i++ {
		go func(index int, waitGroup *sync.WaitGroup) {
			defer waitGroup.Done()
			for j := index; j < int(NumOfVectors); j += MaxCPUs {
				vector := (*points)[j*int(NumOfDimensions) : (j+1)*int(NumOfDimensions)]
				distance := util.CalculateL2Norm(query, &vector, NumOfDimensions)
				distances[j] = distance
			}
		}(i, &wg)
	}

	wg.Wait()
	sort.Float64s(distances[:])
	return distances[:k]
}

func (locator VPTreeKnnLocator) BuildIndex(points *[]int8, numPoints int) *VPTreeNode {
	if numPoints == 0 {
		return nil
	}

	vp := (*points)[(numPoints-1)*int(NumOfDimensions):]
	node := &VPTreeNode{
		vantagePoint: &vp,
	}
	numPoints--

	if numPoints == 0 {
		node.inside = nil
		node.outside = nil
		return node
	}

	distances := make([]float64, numPoints)
	indices := make([]int, numPoints)

	if numPoints >= int(NumOfVectors)/8 {
		numCPUs := 12
		if numPoints <= int(numPoints)/4 {
			numCPUs = 6
		}

		wg := sync.WaitGroup{}
		wg.Add(numCPUs)

		for i := 0; i < numCPUs; i++ {
			go func(index int, waitGroup *sync.WaitGroup) {
				defer waitGroup.Done()
				for j := index; j < numPoints; j += numCPUs {
					point := (*points)[j*int(NumOfDimensions) : (j+1)*int(NumOfDimensions)]
					distances[j] = util.CalculateL2Norm(node.vantagePoint, &point, NumOfDimensions)
					indices[j] = index
				}
			}(i, &wg)
		}
		wg.Wait()
	} else {
		for i := 0; i < numPoints; i++ {
			point := (*points)[i*int(NumOfDimensions) : (i+1)*int(NumOfDimensions)]
			distances[i] = util.CalculateL2Norm(node.vantagePoint, &point, NumOfDimensions)
			indices[i] = i
		}
	}

	sort.Slice(indices, func(i, j int) bool {
		return distances[indices[i]] < distances[indices[j]]
	})

	middle := numPoints / 2
	median := distances[indices[middle]]
	node.radius = median

	insidePoints := make([]int8, 0, middle)
	outsidePoints := make([]int8, 0, numPoints-middle)

	furthestOutsideDistance := -math.MaxFloat64
	furthestInsideDistance := -math.MaxFloat64
	furthestOutsidePointLocation := 0
	furthestInsidePointLocation := 0
	numOutsidePoints := 0
	numInsidePoints := 0

	for i := 0; i < numPoints; i++ {
		point := (*points)[i*int(NumOfDimensions) : (i+1)*int(NumOfDimensions)]
		if distances[i] >= median {
			if distances[i] > furthestOutsideDistance {
				furthestOutsidePointLocation = numOutsidePoints
				furthestOutsideDistance = distances[i]
			}
			outsidePoints = append(outsidePoints, point...)
			numOutsidePoints++
			continue
		}

		if distances[i] > furthestInsideDistance {
			furthestInsidePointLocation = numInsidePoints
			furthestInsideDistance = distances[i]
		}
		insidePoints = append(insidePoints, point...)
		numInsidePoints++
	}

	if numInsidePoints > 0 {
		for l := 0; l < int(NumOfDimensions); l++ {
			temp := insidePoints[furthestInsidePointLocation*int(NumOfDimensions)+l]
			insidePoints[furthestInsidePointLocation*int(NumOfDimensions)+l] =
				insidePoints[(int(numInsidePoints)-1)*int(NumOfDimensions)+l]
			insidePoints[(int(numInsidePoints)-1)*int(NumOfDimensions)+l] = temp
		}
	}

	if numOutsidePoints > 0 {
		for l := 0; l < int(NumOfDimensions); l++ {
			temp := outsidePoints[furthestOutsidePointLocation*int(NumOfDimensions)+l]
			outsidePoints[furthestOutsidePointLocation*int(NumOfDimensions)+l] =
				outsidePoints[(int(numOutsidePoints)-1)*int(NumOfDimensions)+l]
			outsidePoints[(int(numOutsidePoints)-1)*int(NumOfDimensions)+l] = temp
		}
	}

	node.inside = locator.BuildIndex(&insidePoints, len(insidePoints)/int(NumOfDimensions))
	node.outside = locator.BuildIndex(&outsidePoints, len(outsidePoints)/int(NumOfDimensions))

	return node
}

func (locator VPTreeKnnLocator) SearchKNearest(root *VPTreeNode, query *[]int8, k int) []float64 {
	kNearest := make([]int8, 0, (k+1)*int(NumOfDimensions))
	kDistances := make([]float64, 0, k+1)
	queue := make([]*VPTreeNode, 0, NumOfVectors)
	queue = append(queue, root)
	furthestKnnSoFar := math.MaxFloat64

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node == nil {
			continue
		}

		distance := util.CalculateL2Norm(query, node.vantagePoint, NumOfDimensions)
		if distance < furthestKnnSoFar {
			kNearest = append(kNearest, *node.vantagePoint...)
			kDistances = append(kDistances, distance)
			util.SortBasedOn(&kNearest, &kDistances)
			if len(kDistances) > k {
				kDistances = kDistances[:k]
				kNearest = kNearest[:k*int(NumOfDimensions)]
			}
			furthestKnnSoFar = kDistances[len(kDistances)-1]
		}

		if distance < node.radius+furthestKnnSoFar {
			queue = append(queue, node.inside)
		}
		if distance >= node.radius-furthestKnnSoFar {
			queue = append(queue, node.outside)
		}
	}

	return kDistances[:k]
}
