package main

//// #cgo CFLAGS: -g -Wall -O3 -fopenmp
//// #cgo LDFLAGS: -fopenmp
//// #include <stdlib.h>
//// #include "test.c"
//import "C"

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

	start = time.Now()
	res1 := vpTree.SearchKNearest(root, query, Constants.K)
	elapsed = time.Since(start)
	fmt.Println("VPTree kNN: ", elapsed)

	start = time.Now()
	res2 := parallel.SearchKNearest(vectorsInRowMajor, query, Constants.K)
	elapsed = time.Since(start)
	fmt.Println("Parallel kNN: ", elapsed)

	start = time.Now()
	res3 := naive.SearchKNearestNaive(vectorsIn2D, query, Constants.K)
	elapsed = time.Since(start)
	fmt.Println("Naive kNN: ", elapsed)

	for i := 0; i < Constants.K; i++ {
		if res1[i] != res2[i] || res1[i] != res3[i] || res2[i] != res3[i] {
			panic("[ERROR]: K nearest neighbors not consistent amongst methods")
		}
	}

}
