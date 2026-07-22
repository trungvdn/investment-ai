# Workflow Model

**Status:** Draft

**Version:** 2.0

---

# Purpose

Workflow Model định nghĩa cấu trúc và hành vi của một Workflow Definition.

Workflow Definition mô tả quy trình nghiệp vụ nhằm đạt được một mục tiêu cụ thể.

Workflow Definition là một bản thiết kế (Blueprint).

Workflow Definition không đại diện cho quá trình thực thi.

Khi Execution bắt đầu, Execution Layer sẽ tạo Workflow Instance từ Workflow Definition.

---

# Objectives

Workflow Model hướng đến các mục tiêu sau.

- Chuẩn hóa quy trình nghiệp vụ.
- Tách biệt Business Process khỏi Runtime.
- Hỗ trợ tái sử dụng Workflow.
- Hỗ trợ Versioning.
- Cho phép mở rộng Workflow mà không ảnh hưởng Runtime.

---

# Workflow Definition

Workflow Definition mô tả:

- Mục tiêu nghiệp vụ.
- Các Stage cần thực hiện.
- Quan hệ giữa các Stage.
- Điều kiện chuyển tiếp.
- Input.
- Output.

Workflow Definition không lưu:

- Runtime State.
- Runtime Context.
- Runtime Result.

---

# Workflow Structure

Một Workflow Definition bao gồm:

```
Workflow Definition

├── Metadata
├── Input Definition
├── Output Definition
├── Stage Definitions
├── Transition Rules
├── Completion Rules
└── Constraints
```

---

# Workflow Components

## Metadata

Thông tin mô tả Workflow.

Ví dụ:

- Workflow ID
- Name
- Description
- Version
- Owner

---

## Input Definition

Định nghĩa dữ liệu đầu vào.

Ví dụ:

- Stock Symbol
- Portfolio
- User Parameters

---

## Output Definition

Định nghĩa dữ liệu đầu ra.

Ví dụ:

- Recommendation
- Report
- Analysis Result

---

## Stage Definitions

Workflow bao gồm một hoặc nhiều Stage Definition.

Stage mô tả các bước xử lý ở mức nghiệp vụ.

---

## Transition Rules

Quy tắc chuyển giữa các Stage.

Ví dụ:

- Always
- On Success
- On Failure
- Conditional

---

## Completion Rules

Điều kiện kết thúc Workflow.

Ví dụ:

- All Stages Completed
- Target Achieved

---

## Constraints

Các ràng buộc của Workflow.

Ví dụ:

- Maximum Execution Time
- Required Inputs
- Required Capabilities

---

# Workflow Hierarchy

Workflow Definition được tổ chức như sau.

```
Workflow Definition
        │
        ▼
Stage Definition
        │
        ▼
Task Definition
```

Workflow Definition không chứa Runtime Instance.

---

# Workflow Instance

Workflow Definition không được thực thi trực tiếp.

Khi Execution bắt đầu:

```
Workflow Definition

↓

Execution Layer

↓

Workflow Instance
```

Workflow Instance được quản lý bởi Execution Model.

---

# Workflow Relationships

Workflow Definition có quan hệ với các thành phần sau.

## Execution Model

Execution Layer tạo Workflow Instance từ Workflow Definition.

---

## Stage Model

Workflow bao gồm nhiều Stage Definition.

---

## Task Model

Task thuộc về Stage Definition.

---

## Business Capability

Task sẽ tham chiếu đến các Business Capability cần thực hiện.

---

# Workflow Types

Workflow Definition có thể được xây dựng theo nhiều kiểu.

---

## Sequential Workflow

```
Stage A

↓

Stage B

↓

Stage C
```

---

## Parallel Workflow

```
        Workflow

            │

    ┌───────┼────────┐

    ▼       ▼        ▼

Stage A  Stage B  Stage C

    └───────┼────────┘

            ▼

        Next Stage
```

---

## Conditional Workflow

```
Stage A

↓

Condition

├── True → Stage B

└── False → Stage C
```

---

## Hybrid Workflow

Kết hợp Sequential, Parallel và Conditional.

---

# Workflow Boundaries

Workflow Definition chịu trách nhiệm:

- Mô tả quy trình.
- Mô tả Stage.
- Mô tả Transition.
- Mô tả Input và Output.

Workflow Definition không chịu trách nhiệm:

- Runtime State.
- Runtime Context.
- Runtime Scheduling.
- Retry.
- Timeout.
- Monitoring.

Những nội dung này thuộc Execution Layer.

---

# Design Principles

Workflow Definition được xây dựng dựa trên các nguyên tắc sau.

## Declarative

Workflow mô tả "cần làm gì", không mô tả "làm như thế nào".

---

## Business Oriented

Workflow phản ánh quy trình nghiệp vụ.

---

## Reusable

Workflow có thể được sử dụng trong nhiều Execution khác nhau.

---

## Versionable

Workflow có thể tồn tại nhiều phiên bản.

---

## Independent

Workflow không phụ thuộc Runtime.

---

## Extensible

Có thể bổ sung Stage mới mà không ảnh hưởng Execution Engine.

---

# Example

```
Workflow Definition

Analyze Stock

│
├── Stage
│      Data Collection
│
├── Stage
│      Market Analysis
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

Khi người dùng yêu cầu phân tích cổ phiếu:

```
Workflow Definition

↓

Execution Layer

↓

Workflow Instance
```

Workflow Instance sẽ được Execution Layer quản lý trong suốt vòng đời thực thi.

---

# Related Documents

- execution-model.md
- stage-model.md
- task-model.md
- orchestration-model.md
- state-machine.md
