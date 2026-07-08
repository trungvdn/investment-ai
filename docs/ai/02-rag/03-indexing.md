# Chapter 14
# Indexing

---

# Learning Objectives

After completing this chapter, you will understand:

- What indexing is
- Why indexing is critical in RAG
- The document ingestion pipeline
- Document preprocessing
- Cleaning and normalization
- Chunking strategies
- Chunk overlap
- Metadata enrichment
- Embedding generation
- Index construction
- Incremental indexing
- Re-indexing
- Production indexing pipelines

---

# 14.1 Introduction

A Vector Database does not understand documents.

It only understands vectors.

Before documents become searchable, they must pass through a preprocessing pipeline that transforms raw data into indexed knowledge.

This process is called **Indexing**.

```text
Raw Documents

↓

Preprocessing

↓

Chunking

↓

Embedding

↓

Index

↓

Vector Database
```

Indexing is one of the most important stages in a Retrieval-Augmented Generation (RAG) system.

A poor indexing pipeline leads to poor retrieval quality—even if the embedding model and LLM are excellent.

---

# 14.2 What Is Indexing?

Indexing is the process of converting raw knowledge into a searchable representation.

In traditional databases, indexing creates structures such as:

- B-Tree
- Hash Index
- Bitmap Index

These structures accelerate exact lookups.

In RAG, indexing has a broader meaning.

It includes:

- Reading documents
- Cleaning data
- Splitting documents
- Generating embeddings
- Building vector indexes
- Storing metadata

The goal is to make knowledge efficiently retrievable.

---

# 14.3 Why Indexing Matters

Suppose we store the following document.

```
Software Architecture Guide

120 Pages
```

If we embed the entire document as a single vector:

```text
120 Pages

↓

One Embedding
```

A user asks:

```
How does authentication work?
```

The embedding represents the entire document rather than the authentication section.

Retrieval quality becomes poor.

Instead,

```text
120 Pages

↓

Small Chunks

↓

Many Embeddings
```

Now retrieval becomes much more precise.

---

# 14.4 The Indexing Pipeline

A production indexing pipeline typically looks like this.

```text
Knowledge Sources

↓

Document Loader

↓

Cleaning

↓

Normalization

↓

Chunking

↓

Metadata Extraction

↓

Embedding Generation

↓

Vector Index

↓

Vector Database
```

Each stage contributes to the quality of retrieval.

---

# 14.5 Knowledge Sources

Documents may originate from many different systems.

Examples:

- PDF
- Microsoft Word
- Confluence
- Jira
- GitHub
- Markdown
- HTML
- Wiki
- SQL
- REST APIs
- Source Code
- Email
- Slack
- Notion

A production indexing system should support multiple data sources through a common ingestion interface.

---

# 14.6 Document Loading

The first stage reads raw content.

```text
PDF

↓

Parser

↓

Plain Text
```

Similarly,

```text
Markdown

↓

Parser

↓

Plain Text
```

The objective is to extract clean textual content.

---

# 14.7 Cleaning

Raw documents often contain unnecessary information.

Examples:

- HTML tags
- CSS
- JavaScript
- Navigation menus
- Page numbers
- Headers
- Footers
- Duplicate whitespace

Example

Before:

```html
<h1>Authentication</h1>
```

After:

```
Authentication
```

Cleaning removes irrelevant information before embedding.

---

# 14.8 Normalization

Normalization makes documents more consistent.

Typical operations include:

- Unicode normalization
- Whitespace normalization
- Line ending normalization
- Character encoding conversion
- Removing duplicate spaces

Example

Before

```
Authentication     Service
```

After

```
Authentication Service
```

---

# 14.9 Why Chunking Is Necessary

Large Language Models retrieve chunks—not entire books.

Suppose we have:

```
Software Handbook

500 Pages
```

Embedding the entire handbook produces one vector.

Instead,

```text
500 Pages

↓

Chunk 1

Chunk 2

Chunk 3

...

Chunk N
```

Each chunk becomes independently searchable.

Chunking is one of the most important design decisions in RAG.

---

# 14.10 Chunking Strategies

There is no universally correct chunk size.

Different applications require different strategies.

Common approaches include:

- Fixed-size chunking
- Paragraph chunking
- Section chunking
- Semantic chunking
- Recursive chunking
- Code-aware chunking

We will explore each in detail.

---

# 14.11 Fixed-Size Chunking

The simplest strategy splits text every N tokens.

Example

```text
1000 Tokens

↓

200

200

200

200

200
```

Advantages:

- Easy to implement
- Predictable

Disadvantages:

- May split important concepts in half

---

# 14.12 Paragraph Chunking

Each paragraph becomes one chunk.

```text
Paragraph 1

↓

Chunk 1

Paragraph 2

↓

Chunk 2
```

Advantages:

- Preserves logical structure

Disadvantages:

- Paragraph lengths vary significantly

---

# 14.13 Section Chunking

Documents often contain headings.

Example

```
Authentication

Authorization

Deployment

Logging
```

Each section becomes an independent chunk.

This strategy works well for documentation.

---

# 14.14 Semantic Chunking

Instead of splitting by size,

semantic chunking attempts to preserve complete ideas.

Example

```text
Authentication Flow

↓

Entire Chunk
```

rather than

```text
Authentication

↓

Half

↓

Half
```

Semantic chunking generally produces better retrieval quality but requires more sophisticated algorithms.

---

# 14.15 Recursive Chunking

Recursive chunking uses document structure.

Example

```text
Document

↓

Chapter

↓

Section

↓

Paragraph

↓

Sentence
```

The splitter recursively searches for natural boundaries.

This strategy is widely used in production frameworks.

---

# 14.16 Code Chunking

Source code should not be split arbitrarily.

Bad

```go
func Authenticate(
```

↓

Chunk Ends

Good

```go
Entire Function
```

Code chunking should respect:

- Functions
- Classes
- Methods
- Interfaces
- Modules

Developer-focused RAG systems depend heavily on code-aware chunking.

---

# 14.17 Chunk Size

Chunk size affects retrieval quality.

Small chunks

Advantages

- Precise retrieval

Disadvantages

- Loss of context

Large chunks

Advantages

- More context

Disadvantages

- Lower retrieval precision
- Higher token cost

There is always a trade-off.

---

# 14.18 Chunk Overlap

Suppose important information spans two chunks.

Without overlap:

```text
Chunk A

Authentication begins...

------------------

Chunk B

...JWT verification...
```

The relationship is lost.

With overlap:

```text
Chunk A

Authentication begins...

JWT verification

↓

Chunk B

JWT verification...

Authorization
```

Overlap preserves continuity.

Typical overlap ranges from 10% to 30%.

---

# 14.19 Metadata Enrichment

Each chunk should include useful metadata.

Example

```json
{
  "document": "authentication.md",
  "chapter": "JWT",
  "service": "gateway",
  "language": "Go",
  "version": "v3"
}
```

Metadata enables filtering and improves retrieval precision.

---

# 14.20 Embedding Generation

After chunking,

each chunk is embedded independently.

```text
Chunk

↓

Embedding Model

↓

Vector
```

The embedding should capture the semantic meaning of that chunk.

---

# 14.21 Building the Index

The generated vectors are inserted into the Vector Database.

```text
Chunks

↓

Vectors

↓

ANN Index

↓

Vector Database
```

The database constructs its internal search structures automatically.

---

# 14.22 Incremental Indexing

Knowledge evolves continuously.

New documents should be indexed without rebuilding the entire database.

```text
New Document

↓

Chunk

↓

Embedding

↓

Insert
```

This process is called incremental indexing.

---

# 14.23 Re-indexing

Sometimes existing embeddings become outdated.

Examples:

- Better embedding model
- Document updated
- Metadata changed
- Chunking strategy changed

Pipeline

```text
Documents

↓

Reprocess

↓

New Embeddings

↓

Replace Old Index
```

Re-indexing can be expensive but is occasionally necessary.

---

# 14.24 Index Versioning

Production systems often maintain multiple index versions.

```text
Index v1

↓

Index v2

↓

Index v3
```

Benefits:

- Safe rollback
- A/B testing
- Migration
- Evaluation

Never overwrite production indexes directly.

---

# 14.25 Production Indexing Architecture

A production pipeline typically looks like this.

```text
Knowledge Sources

↓

Crawler

↓

Parser

↓

Cleaner

↓

Chunker

↓

Metadata Extractor

↓

Embedding Service

↓

Vector Database

↓

Monitoring
```

Indexing is usually asynchronous.

Users should never wait while documents are being indexed.

---

# 14.26 Index Quality

High-quality indexing requires:

- Clean documents
- Good chunking
- Correct metadata
- High-quality embeddings
- Appropriate chunk overlap
- Consistent preprocessing

Improving indexing quality often yields larger gains than changing the LLM.

---

# 14.27 Common Challenges

### Duplicate Documents

Repeated documents create duplicate retrieval results.

---

### Poor OCR

Scanned PDFs may contain extraction errors.

---

### Broken Formatting

Tables and code blocks may be corrupted during parsing.

---

### Incorrect Chunk Boundaries

Splitting concepts across chunks reduces retrieval quality.

---

### Missing Metadata

Without metadata,

filtering becomes difficult.

---

# 14.28 Production Lessons

Many teams initially focus on selecting the "best" embedding model.

In practice,

retrieval quality often depends more on:

- Document quality
- Chunking strategy
- Metadata
- Index freshness

than on the embedding model itself.

A mediocre embedding model with excellent indexing frequently outperforms a state-of-the-art embedding model with poor indexing.

---

# Production Notes

Think of indexing as preparing a library.

A librarian does not simply place books randomly on shelves.

Instead,

they:

- Organize
- Categorize
- Label
- Index

Only then can readers efficiently find information.

A Vector Database plays the role of the searchable library,

while the indexing pipeline acts as the librarian.

---

# Best Practices

- Clean documents before embedding.
- Preserve semantic boundaries during chunking.
- Use overlap to maintain context.
- Store rich metadata.
- Version production indexes.
- Re-index after major document updates.
- Continuously evaluate retrieval quality.

---

# Common Mistakes

❌ Embedding entire documents as single vectors.

❌ Ignoring document structure.

❌ Using arbitrary chunk sizes.

❌ Forgetting metadata.

❌ Rebuilding the entire index for every update.

❌ Treating indexing as a one-time operation.

---

# Interview Questions

## Basic

1. What is indexing?
2. Why is indexing necessary in RAG?
3. Why do we split documents into chunks?
4. What is metadata?
5. What is chunk overlap?

## Intermediate

6. Compare fixed-size chunking and semantic chunking.
7. Why is code chunking different from text chunking?
8. Explain incremental indexing.
9. Why is document cleaning important?
10. What factors affect retrieval quality?

## Advanced

11. Design an indexing pipeline for an enterprise knowledge base.
12. How would you version production indexes?
13. Why can chunking have a greater impact than the embedding model?
14. How would you index a large source code repository?
15. What metrics would you monitor for an indexing pipeline?

---

# Chapter Summary

In this chapter, you learned:

- What indexing is
- The complete document ingestion pipeline
- Cleaning and normalization
- Chunking strategies
- Chunk overlap
- Metadata enrichment
- Embedding generation
- Index construction
- Incremental indexing
- Re-indexing
- Production indexing architectures

Indexing transforms raw documents into searchable knowledge. It is the foundation of every high-quality RAG system because retrieval quality depends not only on the Vector Database, but also on how knowledge is prepared before it is stored.

In the next chapter, we will explore **Retrieval**, where we examine how user queries are converted into embeddings, how relevant chunks are retrieved, and how modern retrieval strategies maximize the quality of information sent to the LLM.