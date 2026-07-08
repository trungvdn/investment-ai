# Chapter 8
# Embeddings

---

# Learning Objectives

After completing this chapter, you will understand:

- What embeddings are
- Why embeddings are fundamental to modern AI
- How embedding vectors represent meaning
- Semantic similarity
- Embedding spaces
- Vector dimensions
- Distance metrics
- Embedding models
- Sentence embeddings vs Token embeddings
- Embedding pipelines
- Production embedding systems
- The role of embeddings in RAG and AI Agents

---

# 8.1 Introduction

Embeddings are one of the most important concepts in modern Artificial Intelligence.

Every Large Language Model, recommendation system, Retrieval-Augmented Generation (RAG) system, search engine, and AI Agent relies on embeddings.

An embedding converts discrete objects—such as words, sentences, documents, images, or even users—into dense numerical vectors.

These vectors preserve semantic meaning.

Instead of comparing text literally, AI systems compare vectors mathematically.

---

# 8.2 Why Do We Need Embeddings?

Computers understand numbers.

Humans understand language.

Bridging these two worlds requires a mathematical representation.

Example:

```
Dog
```

Token ID:

```
512
```

The number **512** has no semantic meaning.

Instead, the embedding layer converts it into:

```
[-0.41,
 0.82,
 1.15,
 ...
 0.37]
```

Now the neural network can process the information.

---

# 8.3 From Token to Embedding

The complete transformation looks like this.

```text
Text

↓

Tokenizer

↓

Token IDs

↓

Embedding Lookup

↓

Dense Vectors

↓

Transformer
```

Notice that the tokenizer only assigns IDs.

Meaning is learned entirely inside the embedding vectors.

---

# 8.4 What Is an Embedding?

An embedding is a dense vector that captures semantic relationships.

Example

```
Cat

↓

[0.82, -0.15, 0.74, ...]
```

Another word

```
Dog

↓

[0.80, -0.11, 0.69, ...]
```

Because cats and dogs are semantically related, their vectors are close together.

---

# 8.5 Embedding Space

Imagine every embedding as a point inside a very high-dimensional space.

```text
             Tiger

               ●

        Cat ●

Dog ●

                     Lion ●



            Car ●
```

Animals naturally cluster together.

Vehicles appear in another region.

The model learns these relationships automatically during training.

---

# 8.6 Semantic Similarity

Embeddings allow computers to measure semantic similarity.

Example:

```
Doctor

↓

Vector A
```

```
Physician

↓

Vector B
```

Although the words are different, their vectors are very close.

Likewise,

```
Car
```

and

```
Automobile
```

are nearby in embedding space.

This capability enables semantic search.

---

# 8.7 Similarity vs Exact Matching

Traditional search

```
Search:

Car
```

Matches:

```
Car
```

Misses:

```
Automobile
Vehicle
Sedan
SUV
```

Embedding search

```
Car

↓

Embedding

↓

Nearest Neighbors
```

Returns semantically related results.

This is the foundation of modern Retrieval-Augmented Generation.

---

# 8.8 Dense Vectors

Embeddings are called **dense vectors** because almost every dimension contains meaningful values.

Example

```
[0.41,
-0.82,
0.55,
1.29,
...
]
```

This differs from sparse representations such as one-hot encoding.

---

# 8.9 One-Hot Encoding vs Embeddings

Traditional representation

```
Dog

↓

[0,0,0,1,0,0,0]
```

Problems:

- Huge vectors
- No semantic meaning
- Every word is equally distant

Embedding representation

```
Dog

↓

[0.82,
-0.41,
1.15,
...
]
```

Advantages:

- Compact
- Dense
- Semantic
- Learnable

---

# 8.10 Vector Dimensions

Each embedding has a fixed number of dimensions.

Examples

| Model | Dimensions |
|--------|-----------:|
| word2vec | 300 |
| GloVe | 300 |
| BERT | 768 |
| OpenAI text-embedding | 1536+ |
| Modern Embedding Models | 768–4096 |

More dimensions provide greater representational capacity but increase storage and computation.

---

# 8.11 Distance Metrics

To compare embeddings, we calculate the distance between vectors.

Common metrics include:

---

## Cosine Similarity

Measures the angle between vectors.

```
1.0

↓

Exactly the same direction
```

```
0

↓

Unrelated
```

```
-1

↓

Opposite directions
```

Cosine similarity is the most widely used metric in semantic search.

---

## Euclidean Distance

Measures straight-line distance.

```text
Point A

────────────

Point B
```

Smaller distance indicates greater similarity.

---

## Dot Product

Widely used inside Transformer attention mechanisms.

Higher values generally indicate stronger similarity.

---

# 8.12 Embedding Models

Different models produce different embeddings.

Examples

- word2vec
- GloVe
- FastText
- BERT
- E5
- BGE
- OpenAI Embeddings
- Voyage AI
- Qwen Embedding
- Jina Embeddings

Each model is optimized for different applications.

---

# 8.13 Word Embeddings

Earlier embedding models generated vectors for individual words.

Example

```
King

↓

Vector
```

Problems:

```
Bank
```

has multiple meanings.

- River bank
- Financial bank

Traditional word embeddings cannot distinguish them.

---

# 8.14 Contextual Embeddings

Modern Transformers generate context-dependent embeddings.

Sentence 1

```
I deposited money in the bank.
```

Sentence 2

```
The fisherman sat on the bank.
```

Although the word **bank** appears in both sentences, the resulting embeddings differ.

This is one of the major advances introduced by Transformers.

---

# 8.15 Token Embeddings

Inside an LLM, every token receives its own embedding.

Example

```
Artificial

↓

Vector A
```

```
Intelligence

↓

Vector B
```

These vectors continue evolving as they pass through Transformer layers.

The final representation differs significantly from the initial embedding.

---

# 8.16 Sentence Embeddings

Sometimes we need one vector representing an entire sentence.

Example

```
Artificial Intelligence is transforming software engineering.
```

↓

One embedding vector

Sentence embeddings are commonly used for:

- Semantic Search
- RAG
- Clustering
- Recommendation
- Similarity Search

---

# 8.17 Document Embeddings

Entire documents can also be embedded.

```text
Document

↓

Chunking

↓

Embedding Model

↓

Vector Database
```

This process forms the foundation of Retrieval-Augmented Generation.

---

# 8.18 Embedding Pipeline

A production embedding pipeline looks like this.

```text
Document

↓

Cleaning

↓

Chunking

↓

Embedding Model

↓

Vector

↓

Vector Database
```

Later,

```text
Query

↓

Embedding

↓

Similarity Search

↓

Top K Results
```

---

# 8.19 Embeddings in RAG

RAG systems rely heavily on embeddings.

```text
User Question

↓

Embedding

↓

Vector Search

↓

Relevant Documents

↓

LLM
```

The LLM never searches raw text.

It searches vectors.

---

# 8.20 Embeddings in AI Agents

Modern AI Agents use embeddings for:

- Long-term memory
- Episodic memory
- Semantic memory
- Reflection retrieval
- Knowledge retrieval
- Code retrieval
- Similarity search

Example

```text
User Question

↓

Embedding

↓

Memory Retrieval

↓

Relevant Memories

↓

LLM
```

Without embeddings, scalable memory systems would not exist.

---

# 8.21 Embeddings Beyond Text

Embeddings are not limited to language.

They can represent:

- Images
- Audio
- Video
- Source code
- Products
- Users
- Graphs
- Proteins

The same mathematical principles apply.

---

# 8.22 Embedding Quality

A useful embedding should place similar concepts close together.

Example

```
Cat

Dog

Tiger

Lion
```

should cluster together.

Meanwhile,

```
Laptop

Banana

Airplane
```

should appear much farther away.

Embedding quality directly impacts retrieval quality.

---

# 8.23 Production Considerations

Embedding generation is expensive.

Many production systems generate embeddings only once.

```text
Offline

↓

Documents

↓

Embeddings

↓

Vector Database
```

Queries are embedded in real time.

This architecture minimizes latency.

---

# 8.24 Common Challenges

### Poor Chunking

Bad chunks produce poor embeddings.

---

### Wrong Embedding Model

Embedding models trained for general text may perform poorly on source code.

---

### Model Mismatch

Documents and queries should generally use the same embedding model.

---

### Stale Embeddings

When documents change, embeddings must be regenerated.

---

# 8.25 Embedding Lifecycle

```text
Document Created

↓

Chunk

↓

Embedding

↓

Store

↓

Retrieve

↓

Re-rank

↓

LLM

↓

Update Document

↓

Re-embed
```

Embedding management is an ongoing lifecycle rather than a one-time operation.

---

# Production Notes

Embeddings are arguably the most important building block of modern AI Engineering after LLMs themselves.

Nearly every production AI system uses embeddings somewhere in its architecture.

Examples include:

- Vector Databases
- Semantic Search
- Recommendation Systems
- RAG
- AI Memory
- AI Agents
- Code Search
- Knowledge Retrieval

A strong understanding of embeddings is essential before learning Retrieval-Augmented Generation.

---

# Best Practices

- Use the same embedding model for indexing and querying.
- Chunk documents thoughtfully before embedding.
- Regenerate embeddings when documents change.
- Evaluate retrieval quality rather than embedding quality alone.
- Normalize vectors when required by the similarity metric.
- Choose embedding models based on the target domain.

---

# Common Mistakes

❌ Treating token IDs as embeddings.

❌ Assuming larger embedding dimensions always improve retrieval.

❌ Using different embedding models for documents and queries.

❌ Ignoring document chunking.

❌ Expecting embeddings to store factual knowledge.

❌ Forgetting to re-embed updated documents.

---

# Interview Questions

## Basic

1. What is an embedding?
2. Why are embeddings necessary?
3. What is an embedding space?
4. What is semantic similarity?
5. What is the difference between token IDs and embeddings?

## Intermediate

6. Explain cosine similarity.
7. Why are dense vectors better than one-hot encoding?
8. What is the difference between word embeddings and contextual embeddings?
9. Why are sentence embeddings useful for semantic search?
10. How are embeddings used in RAG?

## Advanced

11. Why should queries and documents use the same embedding model?
12. How does chunking affect embedding quality?
13. Why are embeddings fundamental to AI memory systems?
14. How do contextual embeddings solve word ambiguity?
15. Explain the complete embedding lifecycle in a production AI system.

---

# Chapter Summary

In this chapter, you learned:

- What embeddings are
- How embeddings represent semantic meaning
- Embedding spaces
- Semantic similarity
- Dense vectors
- Distance metrics
- Embedding models
- Contextual embeddings
- Sentence and document embeddings
- Embedding pipelines
- Embeddings in RAG
- Embeddings in AI Agents
- Production embedding systems

Embeddings transform language into mathematical representations that preserve semantic relationships. They are the foundation of semantic search, Retrieval-Augmented Generation (RAG), AI memory systems, and intelligent agents.

In the next chapter, we will explore the **Context Window**, examining how Large Language Models process context, why they have memory limitations during inference, and how AI Engineers design systems to overcome those limitations.