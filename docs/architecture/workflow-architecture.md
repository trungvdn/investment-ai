# Workflow Architecture

Version: 2.0

---

# 1. Purpose

Workflow Architecture định nghĩa cách Investment AI điều phối các Capability để hoàn thành một mục tiêu nghiệp vụ.

Workflow là tầng điều phối (Orchestration Layer).

Workflow không chứa Business Logic.

Workflow không phân tích dữ liệu.

Workflow không sinh Recommendation.

Workflow chỉ chịu trách nhiệm:

- Điều phối
- Scheduling
- Dependency Management
- State Management
- Retry
- Timeout
- Observability

Business Logic luôn thuộc Capability.

---

# 2. Design Principles

## Workflow Orchestrates

Workflow chỉ điều phối.

Không chứa thuật toán.

Không chứa Rule.

Không chứa Prompt.

---

## Capability Executes

Capability thực hiện nghiệp vụ.

Workflow không biết implementation.

---

## Knowledge Contract First

Workflow chỉ truyền Knowledge Contracts.

Không truyền DTO nội bộ.

Không truyền Database Model.

---

## Stateless

Workflow không lưu Business State.

Business State thuộc Domain.

Workflow chỉ lưu Execution State.

---

## Observable

Workflow phải quan sát được toàn bộ quá trình thực thi.

---

# 3. Core Concepts

## Workflow

Một quy trình hoàn chỉnh nhằm đạt một Business Goal.

Ví dụ

- Analyze Stock
- Analyze Sector
- Generate Report

Workflow có thể gồm nhiều Task.

---

## Task

Task là một nhóm các Capability cùng phục vụ một mục tiêu trung gian.

Ví dụ

Collect Context

Generate Analyses

Generate Recommendation

Generate Report

Task không chứa Business Logic.

Task có thể chạy Sequential hoặc Parallel.

---

## Capability

Capability là đơn vị nghiệp vụ nhỏ nhất có thể tái sử dụng.

Ví dụ

Financial Analysis

Technical Analysis

Market Intelligence

Valuation

Risk Assessment

Reporting

Capability có Input Contract và Output Contract rõ ràng.

---

## Activity

Activity là các bước thực thi nội bộ của một Capability.

Ví dụ

Validate Input

Load Context

Normalize Data

Generate Evidence

Detect Signals

Analyze

Generate Output

Validate Output

Activity không được Workflow nhìn thấy.

Activity thuộc Capability Implementation.

---

# 4. Architecture Hierarchy

```text
Workflow
    │
    ▼
Task
    │
    ▼
Capability Interface
    │
    ▼
Capability Implementation
    │
    ▼
Activity
```

---

# 5. Execution Model

```text
Workflow Start

↓

Create Workflow Context

↓

Execute Task

↓

Execute Capability

↓

Generate Output Contract

↓

Pass Contract To Next Task

↓

Workflow Complete
```

Workflow không xử lý dữ liệu.

Workflow chỉ chuyển Knowledge Contract.

---

# 6. Workflow Contract

Mỗi Workflow phải khai báo:

Name

Version

Business Goal

Trigger

Input Contract

Output Contract

Task Graph

Retry Policy

Timeout Policy

Execution Policy

Observability Metrics

---

# 7. Task Contract

Task phải định nghĩa:

Task Name

Purpose

Input Contract

Output Contract

Capability Interfaces

Execution Mode

Retry Policy

Timeout

Preconditions

Postconditions

---

# 8. Capability Binding

Workflow không gọi implementation cụ thể.

Workflow chỉ tham chiếu Capability Interface.

Ví dụ

FinancialAnalysis

↓

Go Implementation

hoặc

↓

AI Agent

hoặc

↓

Rule Engine

Workflow không thay đổi.

---

# 9. Activity Lifecycle

Mọi Capability đều thực hiện theo vòng đời chuẩn.

```text
Validate

↓

Load Context

↓

Normalize

↓

Generate Evidence

↓

Detect Signals

↓

Analyze

↓

Generate Output

↓

Validate Output
```

---

# 10. Execution Graph

Workflow được mô hình hóa dưới dạng Directed Acyclic Graph (DAG).

```text
             Task A
             /    \
            ▼      ▼
        Task B   Task C
            \      /
             ▼    ▼
             Task D
```

Node

Task

Edge

Knowledge Contract

---

# 11. Execution Policies

Workflow hỗ trợ:

Sequential

Parallel

Conditional

Fan-Out

Fan-In

Partial Execution

Resume

Cancel

---

# 12. Workflow State Machine

```text
Created

↓

Running

↓

Waiting

↓

Completed
```

Failure States

```text
Retrying

↓

Failed

↓

Cancelled

↓

Timed Out
```

Workflow chỉ quản lý Execution State.

---

# 13. Retry Strategy

Workflow chịu trách nhiệm Retry.

Capability không Retry.

Ví dụ

Timeout

↓

Retry 3 lần

↓

Fallback

↓

Fail

---

# 14. Error Handling

Business Error

Không Retry.

Workflow dừng Task hiện tại.

Technical Error

Retry.

Nếu vẫn thất bại

Workflow Fail.

---

# 15. Knowledge Flow

Workflow không truyền Raw Data trực tiếp giữa các Capability.

Workflow chỉ truyền Knowledge Contracts.

```text
Raw Data

↓

Evidence

↓

Signal

↓

Analysis Result

↓

Investment Thesis

↓

Recommendation

↓

Investment Report
```

---

# 16. Example Workflow

## Analyze Stock

```text
Workflow

│

├── Task
│
│      Collect Context
│
│      ├── Market Intelligence
│      ├── Company Intelligence
│      └── Data Acquisition
│
├── Task
│
│      Generate Analyses
│
│      ├── Financial Analysis
│      ├── Technical Analysis
│      ├── Valuation
│      └── Risk Assessment
│
├── Task
│
│      Generate Recommendation
│
│      └── Investment Decision
│
└── Task
       Generate Report

       └── Reporting
```

---

# 17. Observability

Workflow phải ghi nhận:

Workflow ID

Execution ID

Current Task

Executed Capabilities

Start Time

End Time

Latency

Retry Count

Error Count

Knowledge Contract Version

---

# 18. Future Evolution

Workflow Architecture được thiết kế để hỗ trợ:

- Multi-Agent Execution
- Human-in-the-loop
- Reflection Engine
- Event-Driven Workflow
- Distributed Execution
- Temporal/Cadence Integration
- Kubernetes Job Execution
- MCP Tool Invocation

Không cần thay đổi kiến trúc cốt lõi.
