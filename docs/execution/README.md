# Execution Layer

**Version:** 3.0

---

# Overview

Execution Layer là Runtime Engine chịu trách nhiệm thực thi các Business Process đã được định nghĩa trong hệ thống.

Execution Layer không chứa Business Logic và không định nghĩa Business Process.

Thay vào đó, Execution Layer chuyển các Definition Models thành Runtime Instances và điều phối toàn bộ vòng đời thực thi.

Execution Layer đóng vai trò là cầu nối giữa Business Definition và Infrastructure.

---

# Goals

Execution Layer được xây dựng với các mục tiêu sau.

- Thực thi Business Process.
- Điều phối Runtime Execution.
- Quản lý Context.
- Quản lý State.
- Điều phối Activity.
- Thu thập kết quả.
- Hỗ trợ mở rộng.
- Hỗ trợ Versioning.
- Hỗ trợ Observability.

---

# Architecture

Execution Layer được chia thành ba phần chính.

```
Business Process Definitions

↓

Execution Runtime

↓

Infrastructure
```

Trong đó:

- Business Process Definitions mô tả hệ thống cần làm gì.
- Execution Runtime thực thi các định nghĩa đó.
- Infrastructure cung cấp các dịch vụ kỹ thuật.

---

# Definition Models

Definition Models mô tả Business Process.

```
Workflow Definition

↓

Stage Definition

↓

Task Definition

↓

Activity Definition
```

## Workflow Definition

Định nghĩa một Business Process.

Workflow xác định mục tiêu nghiệp vụ và các Stage cần trải qua.

---

## Stage Definition

Định nghĩa một Business Milestone.

Stage chia Workflow thành các giai đoạn nghiệp vụ.

---

## Task Definition

Định nghĩa một Business Work Unit.

Task mô tả một công việc nghiệp vụ cần hoàn thành để đạt Business Milestone.

---

## Activity Definition

Định nghĩa một Technical Action.

Activity mô tả cách thực hiện một Task.

---

# Runtime Models

Execution Layer chuyển Definition Models thành Runtime Models.

```
Execution

↓

Workflow Instance

↓

Stage Instance

↓

Task Instance

↓

Activity Instance
```

## Execution

Runtime Root của toàn bộ quá trình thực thi.

---

## Workflow Instance

Runtime của Workflow Definition.

---

## Stage Instance

Runtime của Stage Definition.

---

## Task Instance

Runtime của Task Definition.

---

## Activity Instance

Runtime của Activity Definition.

---

# Execution Flow

Execution Layer thực hiện theo trình tự sau.

```
Receive Request

↓

Create Execution

↓

Create Workflow Instance

↓

Create Stage Instance

↓

Create Task Instance

↓

Create Activity Instance

↓

Execute Activities

↓

Collect Results

↓

Complete Task

↓

Complete Stage

↓

Complete Workflow

↓

Complete Execution
```

---

# Responsibilities

Execution Layer chịu trách nhiệm:

- Instantiate Runtime Models.
- Quản lý Context.
- Quản lý Runtime State.
- Điều phối Activity.
- Thu thập kết quả.
- Quản lý vòng đời thực thi.
- Theo dõi Execution.

Execution Layer không chịu trách nhiệm:

- Business Logic.
- Business Rules.
- Business Capability Implementation.
- Business Process Design.

---

# Design Principles

Execution Layer được xây dựng dựa trên các nguyên tắc sau.

## Separation of Definition and Runtime

Definition Models và Runtime Models được tách biệt hoàn toàn.

---

## Business Process Independence

Execution Engine không chứa Business Logic.

---

## Single Responsibility

Mỗi Model chỉ có một trách nhiệm.

---

## Layered Architecture

Business Layer và Runtime Layer được tách biệt rõ ràng.

---

## Composability

Workflow được tạo từ Stage.

Stage được tạo từ Task.

Task được tạo từ Activity.

---

## Extensibility

Có thể mở rộng Workflow, Stage, Task hoặc Activity mà không thay đổi Runtime Engine.

---

# Directory Structure

```
execution/

README.md

specification.md

execution-model.md

workflow-model.md

stage-model.md

task-model.md

activity-model.md

context-model.md

orchestration-model.md

state-machine.md

business-rules.md
```

---

# Document Overview

| Document | Purpose |
|-----------|---------|
| README.md | Tổng quan về Execution Layer |
| specification.md | Phạm vi, mục tiêu và trách nhiệm của Execution Layer |
| execution-model.md | Runtime Execution Model |
| workflow-model.md | Business Process Definition |
| stage-model.md | Business Milestone Definition |
| task-model.md | Business Work Unit Definition |
| activity-model.md | Technical Action Definition |
| context-model.md | Context Management |
| orchestration-model.md | Runtime Orchestration |
| state-machine.md | Runtime State Machine |
| business-rules.md | Runtime Policies và Business Rules áp dụng trong quá trình thực thi |

---

# Related Layers

Execution Layer làm việc cùng các tầng khác trong hệ thống.

```
Presentation Layer

↓

Application Layer

↓

Execution Layer

↓

Business Capabilities

↓

Infrastructure Layer
```

Execution Layer đóng vai trò điều phối, kết nối Business Process với các Business Capability và hạ tầng kỹ thuật.

---

# Summary

Execution Layer là Runtime Engine của hệ thống Investment AI.

Nó chịu trách nhiệm thực thi các Business Process thông qua các Runtime Models, đồng thời đảm bảo sự tách biệt rõ ràng giữa:

- Business Definition
- Runtime Execution
- Technical Infrastructure

Kiến trúc này giúp hệ thống dễ mở rộng, dễ bảo trì, hỗ trợ versioning và tái sử dụng Business Process trên nhiều loại Execution Engine khác nhau.
