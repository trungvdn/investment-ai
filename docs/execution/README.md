# Execution Layer

**Status:** Draft

**Version:** 1.0

---

# Overview

Execution Layer chịu trách nhiệm điều phối việc thực thi các Business Capability nhằm hoàn thành một yêu cầu của người dùng hoặc một quy trình nghiệp vụ.

Execution Layer không chứa Business Logic.

Thay vào đó, Execution Layer quản lý toàn bộ vòng đời của việc thực thi, bao gồm:

- Workflow
- Stage
- Task
- Execution Context
- Execution State
- Capability Invocation
- Result Aggregation

Execution Layer đóng vai trò là cầu nối giữa Presentation Layer và Business Capability Layer.

---

# Objectives

Execution Layer hướng đến các mục tiêu sau:

- Điều phối nhiều Business Capability.
- Chuẩn hóa quy trình thực thi.
- Quản lý trạng thái của quá trình thực thi.
- Quản lý Context xuyên suốt Workflow.
- Cho phép Workflow có thể tái sử dụng.
- Tách biệt Business Logic và Execution Logic.
- Hỗ trợ mở rộng và thay đổi Workflow mà không ảnh hưởng đến Business Capability.

---

# Responsibilities

Execution Layer chịu trách nhiệm:

- Receive Request
- Create Execution
- Build Workflow
- Execute Workflow
- Coordinate Stages
- Coordinate Tasks
- Invoke Business Capabilities
- Manage Execution Context
- Track Execution State
- Aggregate Results
- Return Final Response

Execution Layer không chịu trách nhiệm:

- Acquire Data
- Analyze Market
- Analyze Company
- Calculate Valuation
- Manage Portfolio
- Generate Investment Recommendation

Các nghiệp vụ trên thuộc Business Capability Layer.

---

# Architecture

```
                    User Request
                          │
                          ▼
                 Presentation Layer
                          │
                          ▼
                  Execution Layer
                          │
          ┌───────────────┼───────────────┐
          │               │               │
          ▼               ▼               ▼
     Workflow       Context Manager   State Manager
          │
          ▼
        Stage
          │
          ▼
         Task
          │
          ▼
 Business Capability Layer
          │
          ▼
   Capability Results
          │
          ▼
  Result Aggregation
          │
          ▼
      Final Response
```

---

# Execution Flow

Mỗi User Request được xử lý theo quy trình sau.

```
Receive Request
        │
        ▼
Create Execution
        │
        ▼
Build Workflow
        │
        ▼
Execute Stages
        │
        ▼
Execute Tasks
        │
        ▼
Invoke Business Capabilities
        │
        ▼
Aggregate Results
        │
        ▼
Return Final Response
```

---

# Core Concepts

Execution Layer được xây dựng dựa trên các khái niệm sau.

---

## Execution

Execution đại diện cho toàn bộ vòng đời xử lý của một yêu cầu.

Một Execution bắt đầu khi hệ thống nhận Request và kết thúc khi trả về Final Response.

Một Execution bao gồm một hoặc nhiều Workflow.

---

## Workflow

Workflow mô tả quy trình nghiệp vụ nhằm đạt được một mục tiêu cụ thể.

Ví dụ:

- Analyze Stock
- Screen Opportunities
- Evaluate Portfolio

Một Workflow bao gồm nhiều Stage.

---

## Stage

Stage là một giai đoạn trong Workflow.

Mỗi Stage nhóm các Task có cùng mục tiêu nghiệp vụ.

Ví dụ:

```
Analyze Stock

├── Data Collection
├── Market Analysis
├── Fundamental Analysis
├── Valuation
└── Recommendation
```

Các Stage có thể được thực thi:

- Sequentially
- In Parallel
- Conditionally

---

## Task

Task là đơn vị công việc nhỏ nhất trong Workflow.

Một Task thường thực hiện một chức năng cụ thể như:

- Gọi một Business Capability.
- Chuẩn bị dữ liệu.
- Tổng hợp kết quả.
- Kiểm tra điều kiện.

Task chỉ tập trung vào một mục tiêu duy nhất.

---

## Capability

Capability là đơn vị nghiệp vụ độc lập.

Execution Layer điều phối Capability nhưng không chứa Business Logic của Capability.

Một Capability có thể được tái sử dụng bởi nhiều Workflow khác nhau.

---

## Context

Context lưu trữ thông tin được chia sẻ trong suốt quá trình thực thi.

Ví dụ:

- User Context
- Execution Context
- Workflow Context
- Intermediate Results

Context giúp các Stage và Task trao đổi dữ liệu mà không phụ thuộc trực tiếp vào nhau.

---

## State

State phản ánh trạng thái hiện tại của Execution.

Ví dụ:

- Created
- Running
- Waiting
- Completed
- Failed
- Cancelled

Execution Layer sử dụng State để theo dõi toàn bộ vòng đời của Execution.

---

## Result

Result là đầu ra của:

- Task
- Stage
- Workflow
- Execution

Execution Layer chịu trách nhiệm tổng hợp các Result để tạo Final Response.

---

# Execution Components

Execution Layer bao gồm các thành phần chính.

---

## Workflow Engine

Quản lý việc thực thi Workflow.

---

## Orchestrator

Điều phối Stage và Task.

---

## Context Manager

Quản lý Context xuyên suốt Execution.

---

## State Manager

Theo dõi trạng thái của Execution.

---

## Result Aggregator

Tổng hợp kết quả từ nhiều Capability.

---

# Relationship with Business Capabilities

Execution Layer điều phối Business Capability.

Business Capability không biết mình đang nằm trong Workflow nào.

Ví dụ:

```
Analyze Stock Workflow

│
├── Stage 1 : Data Collection
│      ├── Market Data
│      ├── Financial Data
│      └── News Data
│
├── Stage 2 : Analysis
│      ├── Market Intelligence
│      ├── Fundamental Analysis
│      └── Company Research
│
├── Stage 3 : Valuation
│      └── Valuation
│
└── Stage 4 : Recommendation
       └── Investment Recommendation
```

Nhờ vậy Business Capability có thể được tái sử dụng trong nhiều Workflow khác nhau.

---

# Design Principles

Execution Layer được xây dựng dựa trên các nguyên tắc sau.

## Separation of Concerns

Execution Logic và Business Logic phải tách biệt.

---

## Capability Independence

Business Capability không phụ thuộc Workflow.

---

## Context Driven

Mọi Stage và Task đều sử dụng Execution Context.

---

## State Driven

Execution được quản lý thông qua State.

---

## Workflow Reusability

Workflow có thể tái sử dụng cho nhiều Use Case.

---

## Observability

Toàn bộ Execution phải có khả năng theo dõi và truy vết.

---

## Scalability

Execution phải hỗ trợ mở rộng số lượng Workflow, Stage và Task mà không làm thay đổi Business Capability.

---

# Related Documents

## Foundation

- specification.md
- business-rules.md

---

## Models

- execution-model.md
- workflow-model.md
- stage-model.md
- task-model.md
- context-model.md
- orchestration-model.md
- state-machine.md
