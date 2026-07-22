# Stage Model

**Status:** Draft

**Version:** 2.0

---

# Purpose

Stage Model định nghĩa cấu trúc và hành vi của một Stage Definition.

Stage Definition đại diện cho một giai đoạn (Business Milestone) trong Workflow.

Mỗi Stage tập hợp các Task có cùng mục tiêu nghiệp vụ.

Stage Definition là một phần của Workflow Definition.

Stage Definition không đại diện cho quá trình thực thi.

---

# Objectives

Stage Model hướng đến các mục tiêu sau.

- Chia nhỏ Workflow thành các giai đoạn rõ ràng.
- Nhóm các Task theo mục tiêu nghiệp vụ.
- Tăng khả năng đọc và bảo trì Workflow.
- Cho phép tái sử dụng Stage.
- Hỗ trợ mở rộng Workflow.

---

# Stage Definition

Stage Definition mô tả:

- Mục tiêu của Stage.
- Các Task cần thực hiện.
- Điều kiện bắt đầu.
- Điều kiện hoàn thành.
- Input.
- Output.

Stage Definition không lưu:

- Runtime State.
- Runtime Context.
- Runtime Result.

---

# Stage Structure

Một Stage Definition bao gồm:

```
Stage Definition

├── Metadata
├── Objective
├── Input Definition
├── Output Definition
├── Task Definitions
├── Entry Criteria
├── Exit Criteria
└── Constraints
```

---

# Stage Components

## Metadata

Thông tin mô tả Stage.

Ví dụ:

- Stage ID
- Name
- Description
- Version

---

## Objective

Mục tiêu nghiệp vụ của Stage.

Ví dụ:

- Collect Required Data
- Analyze Market Trend
- Evaluate Financial Health
- Estimate Intrinsic Value

---

## Input Definition

Định nghĩa dữ liệu đầu vào cần thiết.

Ví dụ:

- Stock Symbol
- Market Data
- Financial Statements

---

## Output Definition

Định nghĩa dữ liệu đầu ra của Stage.

Ví dụ:

- Market Analysis
- Financial Analysis
- Valuation Result

---

## Task Definitions

Stage bao gồm một hoặc nhiều Task Definition.

Task là đơn vị công việc nhỏ nhất trong Stage.

---

## Entry Criteria

Điều kiện để Stage được phép bắt đầu.

Ví dụ:

- Previous Stage Completed
- Required Data Available

---

## Exit Criteria

Điều kiện để Stage được xem là hoàn thành.

Ví dụ:

- All Tasks Completed
- Required Output Produced

---

## Constraints

Các ràng buộc của Stage.

Ví dụ:

- Required Tasks
- Required Inputs
- Business Rules

---

# Stage Hierarchy

Stage Definition được tổ chức như sau.

```
Workflow Definition
        │
        ▼
Stage Definition
        │
        ▼
Task Definition
```

Stage Definition không chứa Runtime Instance.

---

# Stage Instance

Stage Definition không được thực thi trực tiếp.

Khi Workflow Instance được tạo:

```
Stage Definition

↓

Execution Layer

↓

Stage Instance
```

Stage Instance được quản lý bởi Execution Model.

---

# Stage Relationships

Stage Definition có quan hệ với các thành phần sau.

## Workflow Model

Stage thuộc về một Workflow Definition.

---

## Task Model

Stage bao gồm nhiều Task Definition.

---

## Business Capability

Task trong Stage sẽ tham chiếu đến Business Capability tương ứng.

---

# Stage Types

Execution Layer không giới hạn loại Stage.

Một số Stage phổ biến:

---

## Data Collection Stage

Thu thập dữ liệu cần thiết.

---

## Validation Stage

Kiểm tra tính hợp lệ của dữ liệu.

---

## Analysis Stage

Thực hiện các phân tích nghiệp vụ.

---

## Decision Stage

Đưa ra đánh giá hoặc quyết định.

---

## Reporting Stage

Tổng hợp và xuất kết quả.

---

# Stage Boundaries

Stage Definition chịu trách nhiệm:

- Mô tả mục tiêu nghiệp vụ.
- Nhóm các Task.
- Định nghĩa Input và Output.
- Xác định điều kiện bắt đầu và kết thúc.

Stage Definition không chịu trách nhiệm:

- Runtime State.
- Runtime Context.
- Scheduling.
- Retry.
- Timeout.
- Monitoring.

Những nội dung này thuộc Execution Layer.

---

# Design Principles

Stage Definition được xây dựng dựa trên các nguyên tắc sau.

## Goal Oriented

Mỗi Stage phục vụ một mục tiêu nghiệp vụ rõ ràng.

---

## Cohesive

Các Task trong cùng một Stage phải liên quan chặt chẽ đến cùng một mục tiêu.

---

## Independent

Stage không chứa Business Logic.

Stage chỉ tổ chức quy trình.

---

## Reusable

Stage có thể được sử dụng trong nhiều Workflow khác nhau.

---

## Extensible

Có thể bổ sung hoặc thay đổi Stage mà không ảnh hưởng đến các Stage khác.

---

# Example

```
Workflow Definition

Analyze Stock

│
├── Stage
│      Data Collection
│      │
│      ├── Acquire Market Data
│      ├── Acquire Financial Data
│      └── Acquire News
│
├── Stage
│      Market Analysis
│      │
│      ├── Analyze Trend
│      ├── Analyze Volume
│      └── Analyze Liquidity
│
├── Stage
│      Fundamental Analysis
│
├── Stage
│      Valuation
│
└── Stage
       Recommendation
```

Khi Workflow được thực thi, mỗi Stage Definition sẽ được Execution Layer tạo thành một Stage Instance và quản lý trong suốt vòng đời thực thi.

---

# Related Documents

- execution-model.md
- workflow-model.md
- task-model.md
- orchestration-model.md
- state-machine.md
