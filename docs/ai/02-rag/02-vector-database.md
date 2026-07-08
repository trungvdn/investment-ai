# Chapter 13
# Vector Databases

---

# Learning Objectives

After completing this chapter, you will understand:

- What a Vector Database is
- Why Vector Databases are required for RAG
- How Vector Databases differ from relational databases
- Dense vectors and vector indexing
- Similarity search
- Approximate Nearest Neighbor (ANN)
- Vector indexing algorithms
- Metadata filtering
- Hybrid retrieval
- Vector database architecture
- Production vector database design

---

# 13.1 Introduction

Large Language Models reason over text.

However, Retrieval-Augmented Generation (RAG) requires searching millions of documents efficiently.

Traditional databases are designed to answer questions such as:

```sql
SELECT *
FROM employees
WHERE id = 100;
```

This is called **exact matching**.

Unfortunately, semantic search is fundamentally different.

Instead of asking:

> Find document ID = 100

we ask:

> Find documents that **mean** something similar to this question.

This requires an entirely different storage and indexing system.

That system is called a **Vector Database**.

---

# 13.2 Why Can't SQL Databases Solve Semantic Search?

Suppose we have the following document.

```
The payment service retries failed transactions.
```

A user asks:

```
How does retry work?
```

There is no keyword:

```
retry
```

inside the document?

Actually there is.

Now consider:

```
How does the system recover failed payments?
```

The keyword **recover** does not exist.

A traditional SQL query cannot understand that:

```
recover

≈

retry
```

A Vector Database can.

---

# 13.3 Exact Search vs Semantic Search

Traditional Database

```text
Query

↓

Exact Match

↓

Result
```

Vector Database

```text
Query

↓

Embedding

↓

Similarity Search

↓

Most Similar Documents
```

The database compares meanings instead of words.

---

# 13.4 What Is a Vector Database?

A Vector Database stores dense embedding vectors and enables efficient similarity search.

Each document is represented as:

```text
Document

↓

Embedding Model

↓

Vector

↓

Vector Database
```

Instead of indexing text directly, the database indexes vectors.

---

# 13.5 Vector Representation

Suppose we have three documents.

```
Document A

↓

[0.23, -0.81, ...]
```

```
Document B

↓

[1.52, 0.18, ...]
```

```
Document C

↓

[-0.41, 0.73, ...]
```

The database stores these vectors.

During retrieval, the user's query is embedded into the same vector space.

---

# 13.6 The Semantic Space

Imagine every embedding as a point in a high-dimensional space.

```text
          Payment

             ●

Refund ●

Retry ●

                      Authentication ●

Database ●
```

Documents discussing similar topics naturally cluster together.

The Vector Database searches nearby vectors.

---

# 13.7 Similarity Search

Given a query vector,

```
Q
```

the Vector Database searches for vectors that are closest.

```text
Query Vector

↓

Nearest Neighbor Search

↓

Top K Results
```

Unlike SQL,

the search does not require exact keywords.

---

# 13.8 Distance Metrics

Similarity search depends on mathematical distance.

Common metrics include:

---

## Cosine Similarity

Measures angle.

```
Higher

↓

More Similar
```

Most widely used.

---

## Euclidean Distance

Measures straight-line distance.

```
Smaller Distance

↓

More Similar
```

---

## Dot Product

Frequently used inside neural networks.

Often performs well when vectors are normalized.

---

# 13.9 Why Brute Force Doesn't Scale

Suppose a company has

```
50 Million Documents
```

For every query,

comparing against all vectors would require:

```
50 Million Distance Calculations
```

per request.

This is far too slow.

Instead,

Vector Databases build indexes.

---

# 13.10 Nearest Neighbor Search

The simplest search algorithm is:

```text
Query

↓

Compare

↓

Every Vector

↓

Sort

↓

Top K
```

This is called

```
Exact Nearest Neighbor Search
```

Advantages

- Perfect accuracy

Disadvantages

- Extremely slow

---

# 13.11 Approximate Nearest Neighbor (ANN)

Production systems usually use

```
Approximate Nearest Neighbor
```

instead.

Idea

```text
Query

↓

Smart Index

↓

Small Candidate Set

↓

Top K
```

Instead of checking every vector,

ANN searches only promising regions.

Advantages

- Much faster

Trade-off

- Slightly lower accuracy

For RAG,

this trade-off is almost always worthwhile.

---

# 13.12 Why ANN Works

Embedding spaces exhibit natural clustering.

```text
Programming

● ● ● ●

Finance

● ● ●

Healthcare

● ● ● ●
```

ANN algorithms exploit these clusters.

Instead of searching the entire space,

they jump directly to relevant neighborhoods.

---

# 13.13 Vector Index

A Vector Index accelerates similarity search.

Without index

```text
Query

↓

Every Vector

↓

Slow
```

With index

```text
Query

↓

Vector Index

↓

Candidates

↓

Fast
```

The index is analogous to a B-tree in relational databases.

---

# 13.14 Common ANN Algorithms

Several indexing algorithms exist.

---

## HNSW

Hierarchical Navigable Small World Graph.

Characteristics

- High accuracy
- Fast search
- Higher memory usage

One of the most popular production algorithms.

---

## IVF

Inverted File Index.

Idea

```text
Vectors

↓

Clusters

↓

Search Cluster

↓

Results
```

Suitable for very large datasets.

---

## PQ

Product Quantization.

Compresses vectors.

Benefits

- Lower memory usage

Trade-offs

- Slight reduction in retrieval quality

Often combined with IVF.

---

# 13.15 Vector Database Architecture

A typical architecture looks like:

```text
Documents

↓

Embedding Model

↓

Vectors

↓

Vector Index

↓

Vector Database

↓

Retriever

↓

LLM
```

The Vector Database acts as the storage layer for semantic retrieval.

---

# 13.16 Metadata

Vectors alone are rarely sufficient.

Each vector is usually associated with metadata.

Example

```json
{
  "title": "payment.md",
  "service": "payment",
  "language": "Go",
  "version": "v2",
  "author": "Backend Team"
}
```

Metadata enables additional filtering.

---

# 13.17 Metadata Filtering

Suppose the query is:

```
Payment retry
```

But we only want:

```
Language = Go
```

Pipeline

```text
Metadata Filter

+

Vector Search

↓

Results
```

This dramatically improves retrieval precision.

---

# 13.18 Top-K Retrieval

Vector Databases rarely return every document.

Instead,

they return only

```
Top K
```

Example

```
Top 5

Top 10

Top 20
```

These documents become the context for the LLM.

Choosing K is an important engineering decision.

---

# 13.19 End-to-End Retrieval

Complete workflow

```text
User Question

↓

Embedding

↓

Vector Database

↓

Top K

↓

Prompt Builder

↓

LLM

↓

Answer
```

Notice that the Vector Database never generates text.

It only retrieves information.

---

# 13.20 Hybrid Search

Semantic search is powerful,

but keyword search is still valuable.

Modern systems often combine both.

```text
Query

↓

Keyword Search

+

Vector Search

↓

Merge

↓

Rank

↓

Top K
```

This is called

```
Hybrid Search
```

It generally outperforms either technique alone.

---

# 13.21 Updating Documents

Knowledge changes over time.

When a document changes,

its embedding must also change.

```text
Document Updated

↓

Generate Embedding

↓

Replace Vector

↓

Reindex
```

Embedding maintenance is a core production responsibility.

---

# 13.22 Popular Vector Databases

Popular production solutions include:

| Database | Type |
|----------|------|
| Pinecone | Managed Cloud |
| Milvus | Open Source |
| Qdrant | Open Source |
| Weaviate | Open Source |
| Chroma | Lightweight |
| pgvector | PostgreSQL Extension |
| Elasticsearch | Hybrid Search |
| OpenSearch | Hybrid Search |

Each system makes different trade-offs regarding scalability, latency, and operational complexity.

---

# 13.23 Choosing a Vector Database

Selection depends on the application.

Small projects

- Chroma
- pgvector

Medium-scale systems

- Qdrant
- Milvus

Large enterprise deployments

- Pinecone
- Weaviate
- Elasticsearch

The best choice depends on:

- Scale
- Cost
- Operational expertise
- Required latency
- Existing infrastructure

---

# 13.24 Production Architecture

A production retrieval system typically looks like:

```text
Knowledge Sources

↓

Document Pipeline

↓

Chunking

↓

Embedding

↓

Vector Database

↓

Retriever

↓

Re-ranker

↓

Context Builder

↓

LLM
```

Notice that the Vector Database is only one component of the retrieval pipeline.

---

# 13.25 Limitations

Vector Databases are not magic.

Challenges include:

- Poor embeddings
- Poor chunking
- Wrong similarity metric
- Duplicate documents
- Stale embeddings
- Large storage requirements
- Approximate retrieval errors

Retrieval quality depends on the entire pipeline, not only the database.

---

# Production Notes

A common misconception is:

> "Using a Vector Database automatically creates a good RAG system."

This is false.

A Vector Database only solves one problem:

> **Efficient semantic retrieval.**

Overall RAG quality depends on:

- Document quality
- Chunking
- Embedding quality
- Metadata
- Retrieval strategy
- Re-ranking
- Prompt construction

The Vector Database is an infrastructure component, not the intelligence itself.

---

# Best Practices

- Use the same embedding model for indexing and querying.
- Store meaningful metadata.
- Combine metadata filtering with vector search.
- Evaluate retrieval quality regularly.
- Rebuild embeddings when documents change.
- Tune Top-K experimentally.
- Consider Hybrid Search for production systems.

---

# Common Mistakes

❌ Treating a Vector Database as a replacement for SQL.

❌ Ignoring metadata filtering.

❌ Retrieving entire documents instead of chunks.

❌ Using different embedding models for indexing and querying.

❌ Forgetting to update embeddings after document changes.

❌ Assuming ANN always returns the optimal result.

---

# Interview Questions

## Basic

1. What is a Vector Database?
2. Why can't SQL databases perform semantic search?
3. What is similarity search?
4. What is an embedding vector?
5. Why do RAG systems require a Vector Database?

## Intermediate

6. Explain Approximate Nearest Neighbor (ANN).
7. Compare Cosine Similarity and Euclidean Distance.
8. What is metadata filtering?
9. Why is Top-K retrieval important?
10. Explain Hybrid Search.

## Advanced

11. Why is ANN preferred over exact search in production?
12. Compare HNSW and IVF.
13. How would you design a scalable Vector Database architecture?
14. Why does retrieval quality depend on more than the Vector Database?
15. What factors influence the choice of a Vector Database for an enterprise system?

---

# Chapter Summary

In this chapter, you learned:

- What a Vector Database is
- Exact search vs semantic search
- Embedding storage
- Similarity search
- Distance metrics
- Approximate Nearest Neighbor (ANN)
- Vector indexing
- Metadata filtering
- Hybrid Search
- Popular Vector Databases
- Production vector database architectures

A Vector Database is the storage and retrieval engine that enables semantic search at scale. It transforms embedding vectors into searchable knowledge, making Retrieval-Augmented Generation practical for real-world AI systems.

In the next chapter, we will explore **Indexing**, examining how documents are transformed into searchable chunks, how indexes are built, and why document preprocessing has a greater impact on retrieval quality than many engineers initially expect.