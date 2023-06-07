package main

import (
	"Weaviate/Constants"
	knn "Weaviate/KNNLocators"
	vm "Weaviate/VectorManager"
	"fmt"
	"time"
)

func main() {
	rowMajorVManager := vm.RandomRowMajorVectorManager{}
	twoDimVManager := vm.Random2DVectorManager{}

	query := rowMajorVManager.GenerateSingleVector()
	vectorsInRowMajor := rowMajorVManager.GenerateVectors()
	vectorsIn2D := twoDimVManager.Transfer2D(vectorsInRowMajor)

	naive := knn.NaiveKnnLocator{}
	parallel := knn.ParallelKnnLocator{}
	vpTree := knn.VPTreeKnnLocator{}

	start := time.Now()
	root := vpTree.BuildIndex(vectorsInRowMajor, int(Constants.NumOfVectors))
	elapsed := time.Since(start)
	fmt.Println("VPT index created: ", elapsed)

	// benchmark indexed query using Vantage Point Tree
	start = time.Now()
	res1 := vpTree.SearchKNearest(root, query, Constants.K)
	elapsed = time.Since(start)
	fmt.Println("VPTree kNN: ", elapsed)

	// benchmark parallel query using goroutines
	start = time.Now()
	res2 := parallel.SearchKNearest(vectorsInRowMajor, query, Constants.K)
	elapsed = time.Since(start)
	fmt.Println("Parallel kNN: ", elapsed)

	// benchmark naive query
	start = time.Now()
	res3 := naive.SearchKNearestNaive(vectorsIn2D, query, Constants.K)
	elapsed = time.Since(start)
	fmt.Println("Naive kNN: ", elapsed)

	//validate results
	for i := 0; i < Constants.K; i++ {
		if res1[i] != res2[i] || res1[i] != res3[i] || res2[i] != res3[i] {
			fmt.Println(res1)
			fmt.Println(res2)
			fmt.Println(res3)
			panic("[ERROR]: K nearest neighbors not consistent amongst methods")
		}
	}
}
