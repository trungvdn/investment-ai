# Chapter 9
# Context Window

---

# Learning Objectives

After completing this chapter, you will understand:

- What a Context Window is
- Why LLMs have context limitations
- Tokens and context length
- Attention complexity
- Context overflow
- Prompt truncation
- Lost in the Middle
- Context Engineering
- Long-context models
- RAG and Context Windows
- Memory and Context Windows
- Best practices for production AI systems

---

# 9.1 Introduction

One of the biggest misconceptions about Large Language Models is that they have unlimited memory.

They do not.

An LLM can only process a limited number of tokens at a time.

This limit is called the **Context Window**.

Think of the Context Window as the model's **working memory** during inference.

Everything inside the window is visible.

Everything outside the window is invisible.

Understanding this limitation is essential for building production AI systems.

---

# 9.2 What Is a Context Window?

A Context Window is the maximum number of tokens an LLM can process in a single inference request.

```text
Prompt

↓

Tokenizer

↓

Context Window

↓

Transformer

↓

Response
```

The context includes **all tokens** involved in the request:

- System Prompt
- User Messages
- Conversation History
- Retrieved Documents
- Tool Outputs
- Generated Response (during decoding)

Every token consumes part of the available context budget.

---

# 9.3 A Simple Analogy

Imagine reading a book through a small window.

```text
+-------------------------+
| Chapter 5               |
|                         |
| Only this page is       |
| visible right now.      |
+-------------------------+
```

You cannot see previous pages unless they are included inside the window.

Similarly, an LLM cannot reason about tokens that lie outside its context window.

Unlike humans, it has no persistent memory of previous pages.

---

# 9.4 How Context Is Constructed

A production AI system usually builds context from multiple sources.

```text
System Prompt

        +

Conversation History

        +

Retrieved Documents

        +

Tool Results

        +

User Question

        ↓

Combined Prompt

        ↓

LLM
```

The Context Builder is responsible for assembling all relevant information before inference.

---

# 9.5 Context Window Sizes

Different models support different context lengths.

| Model Type | Typical Context Window |
|------------|-----------------------:|
| Early GPT Models | 2K–4K |
| Modern Chat Models | 8K–32K |
| Large Context Models | 128K |
| Long Context Models | 256K–1M+ |

Larger context windows enable the model to process more information, but they also increase computational cost.

---

# 9.6 Why Context Windows Are Limited

The primary limitation comes from the **Self-Attention** mechanism.

Every token attends to every other token.

```text
Token 1 ←→ Token 2

Token 1 ←→ Token 3

Token 2 ←→ Token 3

...

Every token connects to every other token.
```

If there are **N** tokens:

```
Attention Complexity

=

O(N²)
```

Doubling the context length results in approximately four times as many attention computations.

This quadratic complexity makes very large context windows computationally expensive.

---

# 9.7 What Happens During Inference?

Suppose a model supports:

```
128K Tokens
```

Prompt:

```text
System Prompt

+

History

+

Documents

+

User Question

=

110K Tokens
```

The model now has only:

```
18K Tokens
```

remaining for generating its response.

The generated output also consumes the context window.

---

# 9.8 Context Overflow

If the input exceeds the maximum context length, one of several things may happen.

### Rejection

```text
Prompt

↓

150K Tokens

↓

Maximum = 128K

↓

Error
```

---

### Truncation

Older tokens may be discarded.

```text
History

↓

Old Messages Removed

↓

Latest Messages Kept
```

---

### Sliding Window

Some systems process large documents in multiple overlapping windows.

```text
Window 1

↓

Window 2

↓

Window 3
```

This technique is commonly used in long-document processing.

---

# 9.9 Context Is Not Memory

Many beginners confuse context with memory.

They are fundamentally different.

### Context

- Temporary
- Exists only during inference
- Disappears after the request

### Memory

- Persistent
- Stored externally
- Retrieved when needed
- Survives across conversations

```text
Conversation

↓

Context Window

↓

LLM

↓

Response

(Context disappears)
```

Memory systems will be explored in later chapters.

---

# 9.10 Attention Distribution

Not every token contributes equally to the final prediction.

Some tokens receive significantly more attention than others.

```text
Question

↓

Attention

↓

Relevant Tokens

↓

Prediction
```

Well-structured prompts improve attention allocation.

---

# 9.11 The "Lost in the Middle" Problem

Research has shown that LLMs often pay more attention to:

- Information at the beginning
- Information at the end

Information located in the middle of long contexts is more likely to be overlooked.

```text
Beginning

★★★★★★★★

Middle

★★★

End

★★★★★★★★
```

This phenomenon is known as the **Lost in the Middle** problem.

It has important implications for prompt design.

---

# 9.12 Context Engineering

Providing more context is **not** always better.

The goal is to provide the **right** context.

Poor Context

```text
100 Pages

↓

Mostly Irrelevant
```

Good Context

```text
5 Relevant Paragraphs

↓

Highly Useful
```

Context Engineering focuses on maximizing the usefulness of every token.

---

# 9.13 Context Compression

Sometimes retrieved information is too large.

Compression techniques include:

- Summarization
- Deduplication
- Keyword extraction
- Ranking
- Reflection

```text
100 Documents

↓

Compress

↓

Top Relevant Information

↓

LLM
```

Compression allows more useful information to fit within the context window.

---

# 9.14 Context Ordering

The order of information also matters.

Recommended ordering:

```text
System Prompt

↓

Instructions

↓

Relevant Memory

↓

Retrieved Knowledge

↓

Conversation History

↓

Current User Question
```

Well-organized context generally produces better responses than randomly ordered information.

---

# 9.15 RAG and Context Windows

Retrieval-Augmented Generation solves the context limitation problem.

Instead of placing an entire knowledge base inside the prompt:

```text
Millions of Documents

↓

Retriever

↓

Top K Documents

↓

Context Window

↓

LLM
```

Only the most relevant documents are inserted into the prompt.

This dramatically improves efficiency.

---

# 9.16 Memory and Context Windows

Persistent memory systems use the same principle.

```text
Thousands of Memories

↓

Memory Retrieval

↓

Top K Memories

↓

Context Window

↓

LLM
```

Only relevant memories are loaded.

Without retrieval, long-term memory would quickly exceed the context window.

---

# 9.17 Tool Calling and Context

External tools also consume context.

Example:

```text
User

↓

LLM

↓

Calculator

↓

Result

↓

Context Window

↓

Final Answer
```

Large tool outputs may require:

- Summarization
- Filtering
- Compression

before being returned to the model.

---

# 9.18 Multi-Turn Conversations

Every conversation grows over time.

```text
User

↓

Assistant

↓

User

↓

Assistant

↓

...
```

Eventually, the conversation exceeds the context window.

Strategies include:

- Removing old messages
- Summarizing history
- Memory retrieval
- Conversation compression

---

# 9.19 Long Context Models

Recent research has produced models supporting very large context windows.

Examples:

- 128K
- 256K
- 1M Tokens

Advantages:

- Long documents
- Large codebases
- Extensive conversations

However, long-context models still face challenges:

- Higher inference cost
- Increased latency
- Lost in the Middle
- Higher memory requirements

A larger context window does not eliminate the need for intelligent retrieval.

---

# 9.20 Context Window in AI Agents

AI Agents rarely send all available information to the LLM.

Instead, they construct context dynamically.

```text
Task

↓

Planner

↓

Retriever

↓

Memory

↓

Prompt Builder

↓

LLM
```

This architecture ensures that only the most relevant information is included.

---

# 9.21 Production Context Pipeline

A modern production AI system typically follows this workflow.

```text
User Question

↓

Memory Retrieval

↓

Knowledge Retrieval

↓

Tool Execution

↓

Context Builder

↓

Prompt Builder

↓

LLM

↓

Response
```

Notice that the LLM is the final consumer of context—not the component responsible for building it.

---

# 9.22 Context Budgeting

Every production system should allocate tokens carefully.

Example:

| Component | Token Budget |
|-----------|-------------:|
| System Prompt | 1K |
| User History | 4K |
| Memory | 2K |
| Retrieved Documents | 6K |
| Tool Results | 2K |
| User Question | 1K |
| Response Budget | 8K |

Total:

```
24K Tokens
```

Budgeting prevents accidental context overflow.

---

# 9.23 Context Engineering vs Prompt Engineering

Many beginners focus exclusively on prompt engineering.

However:

```
Prompt Engineering

↓

How instructions are written.
```

```
Context Engineering

↓

What information is provided.
```

Modern AI systems rely far more on Context Engineering than on clever prompt wording.

For production AI, selecting the right context is usually more valuable than rewriting the prompt.

---

# Production Notes

One of the most important architectural shifts in AI Engineering is recognizing that:

> **LLMs should not be responsible for remembering everything.**

Instead:

- RAG retrieves knowledge.
- Memory retrieves experiences.
- Tools retrieve live data.
- Context Builders assemble relevant information.
- LLMs perform reasoning over that curated context.

This separation of responsibilities enables scalable, efficient, and maintainable AI systems.

---

# Best Practices

- Keep context focused and relevant.
- Retrieve information instead of storing everything in the prompt.
- Budget tokens explicitly.
- Compress long conversations.
- Rank retrieved documents before insertion.
- Place the most important information near the beginning of the prompt.
- Monitor token usage in production.

---

# Common Mistakes

❌ Assuming a larger context window eliminates the need for RAG.

❌ Sending entire documents instead of relevant passages.

❌ Confusing context with long-term memory.

❌ Ignoring context overflow.

❌ Filling the context window with duplicated information.

❌ Forgetting that generated output also consumes the context budget.

---

# Interview Questions

## Basic

1. What is a Context Window?
2. Why do LLMs have context limits?
3. What contributes to the context budget?
4. What happens when the context window is exceeded?
5. Why is context different from memory?

## Intermediate

6. Explain the quadratic complexity of self-attention.
7. What is the Lost in the Middle problem?
8. Why is Context Engineering important?
9. How does RAG address context limitations?
10. Why should retrieved information be ranked?

## Advanced

11. Why doesn't a 1M-token context window eliminate the need for retrieval?
12. How would you design a context budget for an AI Agent?
13. Why is the LLM not responsible for constructing context?
14. How can long conversations be managed efficiently?
15. Explain the relationship between Context Windows, Memory Systems, and RAG.

---

# Chapter Summary

In this chapter, you learned:

- What a Context Window is
- Why context length is limited
- Self-Attention complexity
- Context overflow
- Lost in the Middle
- Context Engineering
- Context compression
- Context ordering
- RAG and Context Windows
- Memory retrieval
- Token budgeting
- Production context pipelines

The Context Window defines the temporary working memory of an LLM during inference. Because this resource is limited, modern AI systems rely on retrieval, memory, context engineering, and intelligent orchestration to provide only the most relevant information to the model.

In the next chapter, we will explore **Inference**, examining how an LLM transforms tokens into generated responses, how decoding strategies influence output quality, and how production inference systems are optimized for latency, throughput, and cost.