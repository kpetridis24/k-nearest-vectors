#include <stdio.h>
#include <stdint.h>
#include <omp.h>

float l2Norm(int *vector1, int *vector2, int size) {
    float sumOfSquares = 0.;

    #pragma omp parallel num_threads(40)
    for (size_t i = 0; i < size; i++) {
        sumOfSquares += (vector1[i] - vector2[i]) * (vector1[i] - vector2[i]);
    }
    return sumOfSquares;
}