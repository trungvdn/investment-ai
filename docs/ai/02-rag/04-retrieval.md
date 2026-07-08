# Chapter 15
# Retrieval

---

# Learning Objectives

After completing this chapter, you will understand:

- What retrieval is
- Why retrieval is the core of every RAG system
- Query understanding
- Query embedding
- Similarity search
- Top-K retrieval
- Metadata filtering
- Query expansion
- Multi-stage retrieval
- Dense retrieval vs sparse retrieval
- Retrieval evaluation
- Production retrieval architectures

---

# 15.1 Introduction

If indexing prepares knowledge for search,

retrieval is the process of finding the right knowledge.

When a user asks a question,

the system must determine:

> **Which documents should the LLM read?**

This may sound simple.

In reality,

retrieval is often the single most important factor affecting RAG quality.

An excellent LLM cannot compensate for poor retrieval.

Conversely,

even a smaller LLM can produce excellent answers when given the correct context.

---

# 15.2 What Is Retrieval?

Retrieval is the process of selecting relevant knowledge from a large collection of documents.

```text
Knowledge Base

↓

Retriever

↓

Relevant Documents

↓

LLM
```

The retriever does **not** generate answers.

Its only responsibility is:

> Find the right information.

---

# 15.3 Why Retrieval Matters

Suppose an enterprise knowledge base contains:

```
5 Million Documents
```

The LLM cannot read everything.

Instead,

the retriever selects:

```
Top 5

or

Top 10
```

documents.

Those documents become the model's temporary knowledge.

This makes retrieval the "eyes" of the LLM.

---

# 15.4 High-Level Retrieval Pipeline

```text
User Question

↓

Query Processing

↓

Embedding

↓

Similarity Search

↓

Metadata Filter

↓

Top K Results

↓

Prompt Builder

↓

LLM
```

Every stage influences the final answer.

---

# 15.5 Query Understanding

Retrieval begins with understanding the user's query.

Example

```
How do I deploy the payment service?
```

The retriever must infer that relevant documents may include:

- Deployment Guide
- Kubernetes Configuration
- Helm Charts
- CI/CD Documentation

rather than simply searching for identical keywords.

---

# 15.6 Query Embedding

The user query is converted into an embedding.

```text
Question

↓

Embedding Model

↓

Query Vector
```

The query embedding must be generated using the same embedding model that was used during indexing.

This ensures both documents and queries occupy the same semantic space.

---

# 15.7 Similarity Search

The query vector is compared with stored document vectors.

```text
Query Vector

↓

Vector Database

↓

Nearest Neighbors

↓

Top K
```

Similarity search retrieves semantically related documents,

even when exact keywords differ.

---

# 15.8 Top-K Retrieval

The retriever rarely returns every matching document.

Instead,

it selects the top **K** candidates.

Example

```
Top 3

Top 5

Top 10
```

Choosing K is a trade-off.

Small K

Advantages

- Lower token usage
- Faster inference

Disadvantages

- Missing useful information

Large K

Advantages

- Higher recall

Disadvantages

- More irrelevant information
- Higher context cost

---

# 15.9 Metadata Filtering

Similarity alone is often insufficient.

Suppose the query is:

```
Authentication
```

But only documents from

```
Version 3
```

are relevant.

Pipeline

```text
Query

↓

Metadata Filter

↓

Similarity Search

↓

Results
```

Metadata filtering improves precision.

---

# 15.10 Dense Retrieval

Dense Retrieval uses embeddings.

```text
Question

↓

Embedding

↓

Vector Search

↓

Relevant Documents
```

Advantages

- Semantic understanding
- Synonym handling
- Robust retrieval

Dense retrieval powers most modern RAG systems.

---

# 15.11 Sparse Retrieval

Sparse retrieval uses keyword statistics.

Examples

- TF-IDF
- BM25

Pipeline

```text
Query

↓

Keyword Matching

↓

Score

↓

Results
```

Advantages

- Fast
- Exact keyword matching

Disadvantages

- Poor semantic understanding

---

# 15.12 Dense vs Sparse Retrieval

| Dense Retrieval | Sparse Retrieval |
|-----------------|------------------|
| Embeddings | Keywords |
| Semantic search | Exact matching |
| Understands synonyms | Does not understand synonyms |
| Better recall | Better precision for exact terms |
| Requires embedding model | No embedding model |

Modern systems often combine both.

---

# 15.13 Hybrid Retrieval

Hybrid Retrieval combines dense and sparse retrieval.

```text
User Query

↓

Dense Search

+

BM25

↓

Merge

↓

Ranking

↓

Top K
```

Advantages

- Better recall
- Better precision
- Handles both semantics and exact keywords

Enterprise search systems frequently adopt this approach.

---

# 15.14 Query Expansion

Users often ask short or ambiguous questions.

Example

```
JWT
```

The retriever may expand it into:

```
JWT

Authentication

Authorization

Access Token
```

Expanded queries increase recall.

Query expansion can be performed using:

- Rules
- Synonym dictionaries
- LLMs
- Historical queries

---

# 15.15 Multi-Query Retrieval

Instead of generating one embedding,

the system generates multiple related queries.

Example

```
Original

↓

Authentication

↓

How authentication works

↓

JWT authentication

↓

Access token validation
```

Each query retrieves documents.

Results are merged.

This technique often improves retrieval quality.

---

# 15.16 Parent-Child Retrieval

Sometimes chunks are too small.

Pipeline

```text
Document

↓

Small Chunks

↓

Retrieve Chunk

↓

Return Parent Section
```

The LLM receives the broader context while retrieval remains precise.

This strategy is especially useful for technical documentation and source code.

---

# 15.17 Multi-Stage Retrieval

Production systems rarely rely on a single retrieval step.

Typical pipeline

```text
Query

↓

Dense Retrieval

↓

Top 100

↓

Metadata Filter

↓

Top 50

↓

Re-ranker

↓

Top 10

↓

LLM
```

Each stage improves result quality.

---

# 15.18 Retrieval Latency

Retrieval must be fast.

Typical latency budget

| Stage | Target |
|--------|-------:|
| Query Embedding | 10–30 ms |
| Vector Search | 10–50 ms |
| Metadata Filter | <10 ms |
| Re-ranking | 20–100 ms |
| Total Retrieval | <150 ms |

Efficient retrieval significantly improves user experience.

---

# 15.19 Recall vs Precision

Two important evaluation metrics are:

### Recall

```
How many relevant documents were retrieved?
```

Higher recall retrieves more useful documents.

---

### Precision

```
How many retrieved documents are actually relevant?
```

Higher precision reduces noise.

An effective retriever balances both.

---

# 15.20 Retrieval Errors

Common retrieval failures include:

### False Positive

Irrelevant document retrieved.

---

### False Negative

Relevant document missed.

False negatives are often more harmful because the LLM never sees the missing knowledge.

---

# 15.21 Retrieval Evaluation

Unlike LLM evaluation,

retrieval evaluation focuses on document quality.

Typical metrics include:

- Recall@K
- Precision@K
- MRR (Mean Reciprocal Rank)
- Hit Rate
- nDCG

Evaluating retrieval independently helps identify bottlenecks before involving the LLM.

---

# 15.22 Retrieval for Source Code

Source code retrieval differs from document retrieval.

Queries may include:

```
Fix nil pointer bug
```

Relevant retrieval targets include:

- Function definitions
- Related interfaces
- Tests
- Historical fixes
- Reflection memories

Developer Agents rely heavily on specialized retrieval strategies.

---

# 15.23 Retrieval in AI Agents

Retrieval extends beyond documents.

An AI Agent may retrieve:

- Knowledge
- Memory
- Source code
- API specifications
- Historical tasks
- Previous conversations

Example

```text
Task

↓

Knowledge Retriever

↓

Memory Retriever

↓

Code Retriever

↓

Context Builder

↓

LLM
```

Different retrievers work together to build a comprehensive context.

---

# 15.24 Production Retrieval Architecture

```text
User Query

↓

Query Understanding

↓

Embedding

↓

Dense Search

↓

Sparse Search

↓

Merge

↓

Metadata Filter

↓

Re-ranker

↓

Context Builder

↓

LLM
```

Modern retrieval systems are multi-stage pipelines rather than single database queries.

---

# 15.25 Common Challenges

### Ambiguous Queries

Example

```
Bank
```

Financial institution?

River bank?

---

### Missing Metadata

Filtering becomes impossible.

---

### Poor Embeddings

Semantic search quality declines.

---

### Poor Chunking

Important information spans multiple chunks.

---

### Wrong Top-K

Too few:

Missing context.

Too many:

Noise.

---

# 15.26 Retrieval Is Not Search Alone

Many beginners think retrieval simply means searching a database.

In reality,

retrieval includes:

- Understanding intent
- Embedding the query
- Searching
- Filtering
- Ranking
- Selecting
- Preparing context

Search is only one component.

---

# Production Notes

One of the biggest architectural lessons in AI Engineering is:

> **The LLM can only reason over what it receives.**

Therefore,

poor retrieval directly limits answer quality.

A state-of-the-art LLM combined with weak retrieval often performs worse than a smaller model with excellent retrieval.

Investing in retrieval quality usually provides one of the highest returns in a production RAG system.

---

# Best Practices

- Use the same embedding model for indexing and querying.
- Combine semantic search with metadata filtering.
- Tune Top-K experimentally.
- Evaluate retrieval independently from generation.
- Monitor latency and retrieval quality.
- Consider Hybrid Retrieval for production deployments.
- Use multi-stage retrieval for large knowledge bases.

---

# Common Mistakes

❌ Treating retrieval as a simple database query.

❌ Returning entire documents instead of focused chunks.

❌ Ignoring metadata.

❌ Using inconsistent embedding models.

❌ Assuming more retrieved documents always improve responses.

❌ Evaluating only the LLM instead of the retrieval pipeline.

---

# Interview Questions

## Basic

1. What is retrieval?
2. Why is retrieval necessary in RAG?
3. What is query embedding?
4. What is Top-K retrieval?
5. What is metadata filtering?

## Intermediate

6. Compare dense retrieval and sparse retrieval.
7. What is Hybrid Retrieval?
8. Explain query expansion.
9. What is parent-child retrieval?
10. Why is retrieval latency important?

## Advanced

11. Design a production retrieval pipeline.
12. Why is retrieval quality often more important than model size?
13. How would you evaluate a retriever independently of the LLM?
14. Explain multi-stage retrieval.
15. How would you design retrieval for an AI Software Engineering Agent?

---

# Chapter Summary

In this chapter, you learned:

- What retrieval is
- Query understanding
- Query embedding
- Similarity search
- Top-K retrieval
- Metadata filtering
- Dense and sparse retrieval
- Hybrid Retrieval
- Query expansion
- Multi-stage retrieval
- Retrieval evaluation
- Production retrieval architectures

Retrieval is the intelligence layer that connects user intent with relevant knowledge. Its purpose is not to answer questions, but to ensure the LLM receives the most useful context for reasoning.

In the next chapter, we will study **Re-ranking**, where retrieved documents are reordered using more sophisticated models to maximize the quality of the final context sent to the LLM.