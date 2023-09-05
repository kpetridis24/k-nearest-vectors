# k-nearest-vectors

Parallel-Indexed multi-dimensional query engine for efficient vector search 

## Motivation

Computing the **distance between high dimensional vectors** is widely applicable in modern technologies. It is used extensively in
vector embeddings databases, like **Weaviate**, when performing queries. Although approximate methods, like **Locality Sensitive Hashing
(LSM)** are useful and make the applications scalable when increase in data occurs, deterministic/brute force implementations are also handy
in smaller datasets.

## Content

I have 3 major implementations in this codebase:

- **Naive approach** to compute the k nearest vectors to a given query
- **Parallel approach** using **Go routines**, parallelizing the naive approach
- **Hybrid Indexed-Parallel approach** using **Vantage Point Tree** structure, optimised for space efficiency and branch pruning with parallelized search across branches
