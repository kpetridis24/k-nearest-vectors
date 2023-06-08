package KNNLocators

import (
	"Weaviate/Constants"
	util "Weaviate/Helpers"
	"math"
	"sort"
	"sync"
)

const (
	NumOfVectors           = Constants.NumOfVectors
	NumOfDimensions        = Constants.NumOfDimensions
	MaxCPUs                = Constants.MaxCPUs
	ThresholdToRunParallel = Constants.ThresholdToRunParallel
	MaximumGoroutines      = Constants.MaximumGoroutines
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
	distances := make([]float64, 0, k)
	localDistances := make([][]float64, Constants.MaxCPUs)
	for i := 0; i < Constants.MaxCPUs; i++ {
		localDistances[i] = make([]float64, 0, Constants.NumOfVectors/Constants.MaxCPUs)
	}

	wg := sync.WaitGroup{}
	wg.Add(Constants.MaxCPUs)

	/*
		Spawning one goroutine per iteration is highly redundant and causes significant
		overhead during the creation and scheduling of the goroutines. This approach
		spawns as many goroutines as available CPUs to optimise usage of resources. Each
		goroutine updates a part of the total array. There are no data races since there
		are no overlapping portions of the array that are assigned to different goroutines.
	*/
	for i := 0; i < Constants.MaxCPUs; i++ {
		go func(index int, waitGroup *sync.WaitGroup, localDistances *[]float64) {
			defer waitGroup.Done()

			for j := index; j < int(Constants.NumOfVectors); j += Constants.MaxCPUs {
				vector := (*points)[j*int(Constants.NumOfDimensions) : (j+1)*int(Constants.NumOfDimensions)]
				distance := util.CalculateL2Norm(query, &vector, Constants.NumOfDimensions)
				*localDistances = append(*localDistances, distance)
			}

			sort.Float64s(*localDistances)
		}(i, &wg, &localDistances[i])
	}

	wg.Wait()
	pointers := make([]uint8, Constants.MaxCPUs)

	/*
		Here we perform a regular merge to get the k smallest values, but instead of
		merging two lists (like in merge sort), we merge Constants.MaxCPUs lists.
	*/
	for len(distances) < k {
		minimumDistance := math.MaxFloat64
		minimumIndex := -1
		for i := 0; i < Constants.MaxCPUs; i++ {
			if localDistances[i][pointers[i]] < minimumDistance {
				minimumDistance = localDistances[i][pointers[i]]
				minimumIndex = i
			}
		}
		distances = append(distances, localDistances[minimumIndex][pointers[minimumIndex]])
		pointers[minimumIndex]++
	}

	return distances
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

	/*
		These values were obtained after manual fine-tuning for 1e-6 vectors of 256 dimensions
		each. Initially, we assign maxCPUs to work in parallel, spawning the same number
		of goroutines. As the number of points decreases after some iterations, we decrease
		the number of CPUs, avoiding letting the introduced overhead to surpass the benefits
		of parallelism. Finally, when number of points are below the threshold, we move on
		with the sequential algorithm, which is more efficient for small number of points.
	*/
	if numPoints >= int(NumOfVectors)/8 {
		numCPUs := MaxCPUs
		if numPoints <= int(numPoints)/4 {
			numCPUs = MaxCPUs / 2
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

	/*
		These variables keep track of the furthest point from our current vantage
		point in the dataset. The reason for this is, selecting the furthest point
		from the current vantage point, as the next vantage point, results in
		noticeable boost in performance. This is because distant vantage points
		split the space in a more balanced manner, designating a clear separation
		between the near neighbors of each, and with all intermediate points.
		Testing and benchmarking indicates that we consistently prune more branches
		while searching, using this technique.
	*/
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

	/*
		Put the furthest point from the current vantage point last, so they become the
		next vantage points.
	*/
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

type NodeDistance struct {
	node     *[]int8
	distance float64
}

func (locator VPTreeKnnLocator) SearchKNearest(root *VPTreeNode, query *[]int8, k int) []float64 {
	kNearest := make([]int8, 0, (k+1)*int(NumOfDimensions))
	kDistances := make([]float64, 0, k+1)
	queue := make([]*VPTreeNode, 0, NumOfVectors)
	queue = append(queue, root)
	furthestKnnSoFar := math.MaxFloat64

	for len(queue) > 0 {
		if len(queue) >= ThresholdToRunParallel {
			nodesToExplore := make(chan *VPTreeNode, 2*MaximumGoroutines)
			localKnn := make(chan NodeDistance, MaximumGoroutines)
			numOfRoutines := int(math.Min(float64(len(queue)), MaximumGoroutines))

			wg := sync.WaitGroup{}
			wg.Add(numOfRoutines)

			for i := 0; i < numOfRoutines; i++ {
				q := i
				go func(furthest float64, waitGroup *sync.WaitGroup,
					nodesToExplore chan<- *VPTreeNode, knnChannel chan<- NodeDistance) {
					node := (queue)[q]

					distance := util.CalculateL2Norm(query, node.vantagePoint, NumOfDimensions)

					if distance < furthest {
						knnChannel <- NodeDistance{node: node.vantagePoint, distance: distance}
					}

					if distance < node.radius+furthest && node.inside != nil {
						nodesToExplore <- node.inside
					}

					if distance >= node.radius-furthest && node.outside != nil {
						nodesToExplore <- node.outside
					}

					waitGroup.Done()
				}(furthestKnnSoFar, &wg, nodesToExplore, localKnn)
			}

			wg.Wait()
			close(nodesToExplore)
			close(localKnn)

			queue = queue[numOfRoutines:]

			for node := range nodesToExplore {
				queue = append(queue, node)
			}

			for pair := range localKnn {
				if pair.distance < furthestKnnSoFar {
					kNearest = append(kNearest, *pair.node...)
					kDistances = append(kDistances, pair.distance)
					util.SortBasedOn(&kNearest, &kDistances)
					if len(kDistances) > k {
						kDistances = kDistances[:k]
						kNearest = kNearest[:k*int(NumOfDimensions)]
					}
					furthestKnnSoFar = kDistances[len(kDistances)-1]
				}
			}
			continue
		}

		node := queue[0]
		queue = queue[1:]

		distance := util.CalculateL2Norm(query, node.vantagePoint, NumOfDimensions)

		if distance < furthestKnnSoFar {
			kNearest = append(kNearest, *node.vantagePoint...)
			kDistances = append(kDistances, distance)
			util.SortBasedOn(&kNearest, &kDistances)
			if len(kDistances) > k {
				kDistances = (kDistances)[:k]
				kNearest = (kNearest)[:k*int(NumOfDimensions)]
			}
			furthestKnnSoFar = (kDistances)[len(kDistances)-1]
		}

		if distance < node.radius+furthestKnnSoFar && node.inside != nil {
			queue = append(queue, node.inside)
		}
		if distance >= node.radius-furthestKnnSoFar && node.outside != nil {
			queue = append(queue, node.outside)
		}
	}

	return kDistances[:k]
}
