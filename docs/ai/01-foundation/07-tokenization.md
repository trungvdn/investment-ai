# Chapter 7
# Tokenization

---

# Learning Objectives

After completing this chapter, you will understand:

- What tokenization is
- Why LLMs cannot understand raw text
- Characters, words, subwords, and tokens
- Modern tokenization algorithms
- Byte Pair Encoding (BPE)
- WordPiece
- SentencePiece
- Byte-level tokenization
- Vocabulary
- Special tokens
- Token counting
- Tokenization in production AI systems

---

# 7.1 Introduction

Humans understand language as words and sentences.

Computers do not.

To a computer, text is simply a sequence of characters.

For example,

```
Hello World
```

is internally represented as

```
72
101
108
108
111
32
87
111
114
108
100
```

However, neural networks cannot directly learn from characters or strings.

Before text can enter a Transformer, it must first be converted into numerical identifiers.

This process is called **Tokenization**.

---

# 7.2 What is Tokenization?

Tokenization is the process of converting text into smaller units called **tokens**.

```text
Raw Text

↓

Tokenizer

↓

Tokens

↓

Token IDs

↓

Embeddings

↓

Transformer
```

A tokenizer acts as the bridge between human language and neural networks.

Without tokenization, Large Language Models cannot process text.

---

# 7.3 Why Not Use Characters?

One possible solution would be to process text character by character.

Example:

```
ChatGPT
```

↓

```
C

h

a

t

G

P

T
```

Although simple, character-level tokenization has major disadvantages.

### Longer Sequences

Sentence:

```
Artificial Intelligence
```

Character tokens:

```
A r t i f i c i a l ...

```

More tokens mean:

- Higher computation cost
- More memory usage
- Longer inference time

---

### Difficult Learning

The model must first learn:

```
Characters

↓

Words

↓

Meaning
```

Learning becomes unnecessarily difficult.

---

# 7.4 Why Not Use Whole Words?

Another idea is to tokenize by words.

Example

```
Artificial Intelligence changes software.
```

↓

```
Artificial

Intelligence

changes

software
```

Advantages:

- Shorter sequences
- Easier interpretation

Problems:

Every language contains millions of words.

New words appear constantly.

Examples:

- ChatGPT
- Kubernetes
- TensorRT
- DeepSeek
- GitHub Copilot

A fixed vocabulary of words quickly becomes impractical.

---

# 7.5 Subword Tokenization

Modern LLMs use **subword tokenization**.

Instead of characters or complete words, the tokenizer learns common fragments.

Example

```
unbelievable
```

↓

```
un

believ

able
```

Benefits:

- Smaller vocabulary
- Better handling of rare words
- Efficient token usage
- Generalization to unseen words

Most modern LLMs use some form of subword tokenization.

---

# 7.6 The Tokenization Pipeline

The complete preprocessing pipeline looks like this.

```text
Raw Text

↓

Normalize

↓

Split into Subwords

↓

Convert to Token IDs

↓

Embedding Layer

↓

Transformer
```

Each stage prepares the input for the neural network.

---

# 7.7 Vocabulary

Every tokenizer owns a fixed vocabulary.

Example

| Token | ID |
|--------|---:|
| hello | 125 |
| world | 981 |
| AI | 512 |
| software | 734 |

The vocabulary maps every token to a unique integer.

The neural network only sees these integer IDs.

---

# 7.8 Token IDs

Example:

Sentence:

```
Hello AI
```

Tokenizer output:

```
Hello

↓

Token ID = 1543

AI

↓

Token ID = 932
```

Final input:

```
[1543, 932]
```

These IDs are later transformed into embedding vectors.

---

# 7.9 Embeddings

The tokenizer itself does **not** understand meaning.

It simply converts text into IDs.

Meaning is learned by the embedding layer.

```text
Token IDs

↓

Embedding Lookup

↓

Dense Vectors

↓

Transformer
```

Example:

```
1543

↓

[0.21, -0.83, 1.25, ...]
```

This vector is what the Transformer actually processes.

---

# 7.10 Byte Pair Encoding (BPE)

Byte Pair Encoding (BPE) is one of the most influential tokenization algorithms.

Originally developed for data compression, it was later adapted for NLP.

### Basic Idea

Start with characters.

```
l

o

w

e

r
```

Repeatedly merge the most frequent adjacent pairs.

Example

```
l + o

↓

lo
```

Then

```
lo + w

↓

low
```

Eventually, common words become single tokens.

Rare words remain combinations of smaller pieces.

---

## Example

Word:

```
lowest
```

May become

```
low

est
```

Another word

```
lower
```

becomes

```
low

er
```

The tokenizer efficiently reuses learned subwords.

---

# 7.11 WordPiece

WordPiece is another subword algorithm.

It was introduced by Google and became popular through BERT.

Example

```
playing
```

↓

```
play

##ing
```

The prefix

```
##
```

indicates that the token continues the previous token.

---

# 7.12 SentencePiece

SentencePiece was developed by Google.

Unlike BPE and WordPiece, it operates directly on raw text.

It does not require whitespace.

Example:

```
Machine Learning
```

↓

```
▁Machine

▁Learning
```

The symbol

```
▁
```

represents a space.

SentencePiece works well for multilingual languages where whitespace rules vary.

---

# 7.13 Byte-Level Tokenization

Some modern models tokenize bytes rather than Unicode characters.

Advantages:

- Supports every language
- Supports emojis
- Supports source code
- No unknown characters

Models such as GPT-2 use byte-level BPE.

This allows robust handling of arbitrary input.

---

# 7.14 Special Tokens

Every LLM reserves special tokens for specific purposes.

Examples:

| Token | Purpose |
|--------|---------|
| `<BOS>` | Beginning of sequence |
| `<EOS>` | End of sequence |
| `<PAD>` | Padding |
| `<UNK>` | Unknown token |
| `<CLS>` | Classification |
| `<SEP>` | Separator |

Different models define different special tokens.

---

# 7.15 Tokenization Example

Sentence

```
ChatGPT helps software engineers.
```

Possible tokenization

```
Chat

GPT

helps

software

engineers

.
```

Token IDs

```
412

875

152

934

684

19
```

Embedding vectors

```
↓

[vector]

↓

Transformer
```

Notice that tokens do **not** necessarily correspond to words.

---

# 7.16 Why Token Counts Matter

LLMs charge and operate based on **tokens**, not characters.

Example

```
1000 words

↓

Approximately

1300~1500 tokens
```

The exact number depends on:

- Language
- Vocabulary
- Punctuation
- Numbers
- Source code

For AI Engineers, token count directly affects:

- Cost
- Latency
- Context window usage

---

# 7.17 Tokenization Across Languages

Different languages produce different token counts.

Example

English

```
Artificial Intelligence
```

↓

2–4 tokens

Japanese

```
人工知能
```

↓

Different token count

Chinese

```
人工智能
```

↓

Different token count

Vietnamese

```
Trí tuệ nhân tạo
```

↓

Often more tokens than expected

The tokenizer's vocabulary largely determines efficiency.

---

# 7.18 Source Code Tokenization

Programming languages are also tokenized.

Example

```go
func add(a int, b int) int {
    return a + b
}
```

Possible tokens

```
func

add

(

a

int

,

...

}
```

Modern code models learn programming languages using the same tokenization principles as natural language.

---

# 7.19 Tokenization in LLM Inference

Complete inference pipeline

```text
Prompt

↓

Tokenizer

↓

Token IDs

↓

Embeddings

↓

Transformer

↓

Logits

↓

Next Token

↓

Detokenizer

↓

Response
```

The tokenizer appears both:

- Before inference
- After inference

The output token IDs are converted back into readable text.

---

# 7.20 Production Considerations

Tokenization significantly impacts production systems.

Long prompts increase:

- GPU memory usage
- Inference latency
- API cost

AI Engineers often optimize prompts by reducing unnecessary tokens.

Example

Instead of:

```
Please kindly answer the following question in as much detail as possible.
```

Use

```
Answer:
```

The second prompt consumes far fewer tokens while conveying nearly the same instruction.

---

# 7.21 Tokenization and Context Window

Every model has a maximum context window.

Example

```
128K Tokens
```

If the prompt exceeds this limit:

```text
Input

↓

Tokenizer

↓

129K Tokens

↓

Too Large

↓

Rejected
```

or

```
Old tokens are truncated.
```

Managing token budgets is one of the most important responsibilities in AI Engineering.

---

# 7.22 Tokenization in AI Engineering

Most AI Engineers never implement tokenization algorithms.

Instead, they work with existing tokenizers provided by model vendors.

Typical workflow:

```text
User Input

↓

Tokenizer

↓

Context Builder

↓

Prompt Builder

↓

LLM

↓

Output
```

However, understanding tokenization is essential because it affects:

- Prompt engineering
- Context engineering
- Cost optimization
- Retrieval-Augmented Generation
- Memory systems
- AI Agent performance

---

# Production Notes

Many beginners assume that LLMs "read words."

They do not.

LLMs process **tokens**, which are numerical representations of subword units.

Every optimization involving prompts, RAG, memory, or agent workflows ultimately depends on efficient token utilization.

For production AI systems, managing tokens is as important as managing CPU or memory in traditional software systems.

---

# Best Practices

- Think in tokens rather than words.
- Keep prompts concise.
- Measure token usage during development.
- Be aware that different models use different tokenizers.
- Budget tokens for both input and output.
- Design retrieval systems that return only relevant information.

---

# Common Mistakes

❌ Assuming one word always equals one token.

❌ Ignoring token costs when designing prompts.

❌ Forgetting that source code consumes many tokens.

❌ Assuming every LLM shares the same tokenizer.

❌ Filling the context window with unnecessary information.

---

# Interview Questions

## Basic

1. What is tokenization?
2. Why can't LLMs process raw text directly?
3. What is a token?
4. What is a vocabulary?
5. What is a token ID?

## Intermediate

6. Why do modern LLMs use subword tokenization?
7. Explain Byte Pair Encoding (BPE).
8. What is the difference between WordPiece and SentencePiece?
9. Why are embeddings separate from tokenization?
10. What are special tokens used for?

## Advanced

11. Why do different languages produce different token counts?
12. How does tokenization affect inference cost?
13. Why is token budgeting important in RAG systems?
14. How does tokenization influence context window utilization?
15. Why is tokenization considered a critical component of production AI systems?

---

# Chapter Summary

In this chapter, you learned:

- What tokenization is
- Why tokenization is necessary
- Characters, words, subwords, and tokens
- Vocabulary and token IDs
- Byte Pair Encoding (BPE)
- WordPiece
- SentencePiece
- Byte-level tokenization
- Special tokens
- Token counting
- Tokenization in production AI systems

Tokenization is the first step in every LLM inference pipeline. It converts human language into numerical representations that neural networks can process efficiently. In the next chapter, we will explore **Embeddings**, where these token IDs are transformed into dense vectors that capture semantic meaning and become the true input to the Transformer.