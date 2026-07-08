# Chapter 10
# LLM Inference

---

# Learning Objectives

After completing this chapter, you will understand:

- What inference is
- Training vs inference
- The complete LLM inference pipeline
- KV Cache
- Autoregressive generation
- Logits
- Softmax
- Decoding strategies
- Temperature
- Top-k Sampling
- Top-p Sampling
- Beam Search
- Streaming inference
- Batch inference
- Production inference optimization

---

# 10.1 Introduction

When people interact with ChatGPT, Claude, Gemini, or other Large Language Models, they are **not training the model**.

They are performing **inference**.

Inference is the process of using a trained model to generate predictions from new inputs.

For an LLM, inference means:

> Given a prompt, predict the next token repeatedly until the response is complete.

Although this sounds simple, modern LLM inference is one of the most computationally intensive workloads in AI.

---

# 10.2 Training vs Inference

Training and inference are fundamentally different phases.

```text
               Training

Dataset
    │
    ▼
Transformer
    │
    ▼
Prediction
    │
    ▼
Loss
    │
    ▼
Backpropagation
    │
    ▼
Update Parameters

-------------------------------------

              Inference

Prompt
    │
    ▼
Transformer
    │
    ▼
Prediction
    │
    ▼
Response
```

| Training | Inference |
|-----------|-----------|
| Learns parameters | Uses learned parameters |
| Updates weights | Weights never change |
| Very expensive | Relatively cheaper |
| Offline | Online |
| Requires labels | No labels required |

---

# 10.3 High-Level Inference Pipeline

Every LLM follows approximately the same pipeline.

```text
User Prompt

↓

Tokenizer

↓

Token IDs

↓

Embedding

↓

Transformer

↓

Logits

↓

Sampling

↓

Next Token

↓

Append Token

↓

Repeat

↓

Final Response
```

Everything after tokenization happens numerically.

---

# 10.4 Step 1 — Prompt Processing

The process begins with user input.

Example

```
Explain what AI is.
```

The application sends the prompt to the tokenizer.

---

# 10.5 Step 2 — Tokenization

The tokenizer converts text into token IDs.

Example

```
Explain

↓

3128
```

```
AI

↓

982
```

```
is

↓

51
```

The model never receives strings.

It receives integers.

---

# 10.6 Step 3 — Embedding Lookup

Each token ID becomes an embedding vector.

```text
3128

↓

Embedding Vector
```

The Transformer processes vectors rather than integers.

---

# 10.7 Step 4 — Transformer Forward Pass

The embeddings pass through multiple Transformer layers.

```text
Embedding

↓

Transformer Layer

↓

Transformer Layer

↓

Transformer Layer

↓

...

↓

Final Hidden State
```

Each layer refines the contextual representation.

---

# 10.8 Step 5 — Logits

The final Transformer layer produces **logits**.

A logit is an unnormalized score for every token in the vocabulary.

Example

| Token | Logit |
|--------|------:|
| cat | 8.7 |
| dog | 7.1 |
| car | 1.5 |
| banana | -2.8 |

These values are **not probabilities**.

---

# 10.9 Step 6 — Softmax

Softmax converts logits into probabilities.

```text
Logits

↓

Softmax

↓

Probability Distribution
```

Example

| Token | Probability |
|--------|------------:|
| cat | 71% |
| dog | 20% |
| car | 7% |
| banana | 2% |

The probabilities always sum to:

```
100%
```

---

# 10.10 Step 7 — Token Selection

The model must choose one token.

Example

```
Prompt

↓

Probability Distribution

↓

Select Token

↓

Append

↓

Repeat
```

How the token is selected depends on the decoding strategy.

---

# 10.11 Autoregressive Generation

Modern LLMs generate text one token at a time.

Example

```
The

↓

The cat

↓

The cat sat

↓

The cat sat on

↓

The cat sat on the

↓

...
```

Each newly generated token becomes part of the next input.

This process is called **autoregressive generation**.

---

# 10.12 Stopping Criteria

Generation continues until one of several conditions occurs.

Examples:

- End-of-sequence token
- Maximum output length
- Stop sequence
- User interruption

```text
Generate

↓

EOS?

↓

Yes

↓

Stop
```

---

# 10.13 Greedy Decoding

The simplest strategy always chooses the most probable token.

Example

```
Apple

↓

100%

↓

Selected
```

Advantages

- Fast
- Deterministic

Disadvantages

- Repetitive
- Less creative

---

# 10.14 Temperature

Temperature controls randomness.

Low temperature

```
0.1
```

Characteristics

- Deterministic
- Conservative
- Stable

High temperature

```
1.5
```

Characteristics

- Creative
- Diverse
- Less predictable

Typical production values

```
0.2 ~ 0.8
```

---

# 10.15 Top-k Sampling

Instead of considering the entire vocabulary, Top-k only considers the top **k** candidates.

Example

```
Vocabulary

↓

Top 50 Tokens

↓

Random Selection
```

Advantages

- More diverse
- Prevents unlikely words

---

# 10.16 Top-p (Nucleus Sampling)

Top-p selects the smallest set of tokens whose cumulative probability exceeds a threshold.

Example

```
cat      45%

dog      25%

bird     15%

fish     5%

...

↓

Top-p = 90%

↓

First three tokens
```

This strategy adapts dynamically to different probability distributions.

Top-p is widely used in production LLMs.

---

# 10.17 Beam Search

Beam Search keeps multiple candidate sequences simultaneously.

```text
Start

↓

Candidate A

Candidate B

Candidate C

↓

Expand

↓

Best Sequence
```

Beam Search is common in:

- Translation
- Speech Recognition

It is less common for conversational LLMs because responses may become overly deterministic.

---

# 10.18 Streaming Inference

Modern AI assistants stream tokens as soon as they are generated.

Instead of waiting for the complete response:

```
Artificial

↓

Artificial Intelligence

↓

Artificial Intelligence is

↓

...
```

Users receive partial responses immediately.

Benefits:

- Better user experience
- Lower perceived latency

---

# 10.19 Batch Inference

Inference servers often process multiple requests together.

```text
Request A

Request B

Request C

↓

Batch

↓

GPU

↓

Responses
```

Advantages

- Better GPU utilization
- Higher throughput
- Lower cost per request

---

# 10.20 KV Cache

One of the most important optimizations in LLM inference is the **Key-Value Cache (KV Cache)**.

Without KV Cache

```
Token 1

↓

Recompute Everything

↓

Token 2

↓

Recompute Everything

↓

Token 3
```

Very inefficient.

---

With KV Cache

```text
Token 1

↓

Store Keys & Values

↓

Token 2

↓

Reuse Cache

↓

Token 3

↓

Reuse Cache
```

Only the new token needs to be processed.

Benefits:

- Dramatically faster inference
- Lower GPU computation
- Lower latency

Almost every production LLM server uses KV Cache.

---

# 10.21 Why Inference Is Expensive

Inference may appear simple, but each generated token requires:

- Matrix multiplications
- Attention computation
- Memory access
- Sampling

For a 70B parameter model, generating hundreds of tokens requires billions of floating-point operations.

GPU memory bandwidth is often the primary bottleneck.

---

# 10.22 Quantization

Large models require significant memory.

Quantization reduces precision.

Example

```
FP32

↓

FP16

↓

INT8

↓

INT4
```

Benefits

- Lower memory usage
- Faster inference
- Lower cost

Trade-offs

- Slight quality reduction

---

# 10.23 Continuous Batching

Modern inference engines continuously combine incoming requests.

```text
User A

User B

User C

↓

Dynamic Batch

↓

GPU

↓

Responses
```

Continuous batching significantly increases throughput.

Frameworks such as vLLM rely heavily on this optimization.

---

# 10.24 Speculative Decoding

Speculative decoding accelerates generation by using a smaller model to predict candidate tokens.

```text
Small Model

↓

Draft Tokens

↓

Large Model Verification

↓

Accept / Reject
```

Benefits

- Lower latency
- Higher throughput

This technique is increasingly common in production systems.

---

# 10.25 Inference Architecture

A modern inference system typically looks like this.

```text
User

↓

API Gateway

↓

Load Balancer

↓

Inference Server

↓

Tokenizer

↓

LLM Runtime

↓

GPU

↓

Streaming Response
```

The application rarely communicates directly with the GPU.

---

# 10.26 Inference Frameworks

Popular inference frameworks include:

- vLLM
- TensorRT-LLM
- Ollama
- llama.cpp
- Hugging Face Text Generation Inference (TGI)
- SGLang

These frameworks provide:

- KV Cache management
- Quantization
- Continuous batching
- Streaming
- Multi-GPU inference

---

# 10.27 Inference Metrics

Production systems monitor several key metrics.

### Latency

Time required to produce a response.

---

### Throughput

Number of requests processed per second.

---

### Tokens Per Second (TPS)

Generation speed.

---

### Time To First Token (TTFT)

Time before the first token appears.

Critical for interactive chat applications.

---

### GPU Utilization

Measures hardware efficiency.

---

# 10.28 AI Engineer Perspective

Most AI Engineers never implement Transformer inference manually.

Instead, they optimize:

- Prompt quality
- Context construction
- Retrieval
- Memory
- Tool execution
- Response quality
- Latency
- Cost

Understanding inference allows AI Engineers to make informed architectural decisions.

---

# Production Notes

Inference—not training—is the dominant operational cost for most AI applications.

Optimizing inference often has a greater business impact than improving model accuracy.

Modern production systems focus on:

- KV Cache
- Continuous batching
- Quantization
- Efficient context construction
- Streaming responses
- Intelligent routing

These optimizations enable high-performance, cost-effective AI services.

---

# Best Practices

- Stream responses whenever possible.
- Reuse KV Cache efficiently.
- Monitor TTFT and TPS.
- Minimize unnecessary context.
- Use quantized models when appropriate.
- Select decoding strategies based on application requirements.

---

# Common Mistakes

❌ Confusing training with inference.

❌ Assuming inference simply retrieves memorized answers.

❌ Ignoring context size during inference.

❌ Using unnecessarily high temperatures in production.

❌ Forgetting that every generated token requires another forward pass.

❌ Underestimating GPU memory bandwidth constraints.

---

# Interview Questions

## Basic

1. What is inference?
2. How does inference differ from training?
3. What are logits?
4. What does Softmax do?
5. What is autoregressive generation?

## Intermediate

6. Explain Top-k Sampling.
7. Explain Top-p Sampling.
8. What is the purpose of Temperature?
9. Why is KV Cache important?
10. Why is streaming inference beneficial?

## Advanced

11. Why is inference expensive for Large Language Models?
12. Explain continuous batching.
13. What is speculative decoding?
14. How does quantization improve inference performance?
15. Design a high-level production inference architecture for an AI application.

---

# Chapter Summary

In this chapter, you learned:

- The complete LLM inference pipeline
- Training vs inference
- Logits and Softmax
- Autoregressive generation
- Decoding strategies
- Temperature
- Top-k and Top-p Sampling
- Beam Search
- Streaming inference
- KV Cache
- Quantization
- Continuous batching
- Production inference architectures

Inference is the operational heart of every AI application. While training creates the model, inference delivers intelligence to users in real time. Understanding inference enables AI Engineers to optimize latency, throughput, scalability, and cost.

With this chapter, you have completed the **Foundation** section of the handbook. You now understand the complete journey from Artificial Intelligence → Machine Learning → Deep Learning → Neural Networks → Transformers → Large Language Models → Tokenization → Embeddings → Context Windows → Inference. These concepts provide the foundation for the next major section: **Retrieval-Augmented Generation (RAG)**, where we extend LLMs with external knowledge and build production-ready AI systems.