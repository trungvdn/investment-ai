# Business Rules

**Status:** Draft

**Version:** 1.0

---

# Purpose

Business Rules định nghĩa các quy tắc nghiệp vụ chi phối việc thực thi Business Process.

Business Rules xác định điều kiện để Runtime có thể tiếp tục, hoàn thành hoặc kết thúc quá trình thực thi.

Business Rules không mô tả cách thực thi kỹ thuật.

---

# Objectives

Business Rules hướng đến các mục tiêu sau.

- Đảm bảo Business Process luôn tuân thủ nghiệp vụ.
- Chuẩn hóa các điều kiện thực thi.
- Tách biệt Business Logic khỏi Execution Engine.
- Cho phép thay đổi nghiệp vụ mà không ảnh hưởng đến Runtime.

---

# Scope

Business Rules áp dụng cho toàn bộ Definition Models.

```
Workflow

↓

Stage

↓

Task

↓

Activity
```

Mỗi Definition có thể định nghĩa các Business Rules riêng.

---

# Responsibilities

Business Rules chịu trách nhiệm:

- Định nghĩa điều kiện thực thi.
- Định nghĩa điều kiện hoàn thành.
- Định nghĩa điều kiện thất bại.
- Kiểm tra tính hợp lệ của dữ liệu nghiệp vụ.

Business Rules không chịu trách nhiệm:

- Runtime Scheduling.
- State Transition.
- Context Management.
- Activity Execution.
- Retry Policy.
- Timeout Policy.

---

# Rule Categories

## Entry Rules

Điều kiện để Runtime được phép bắt đầu.

Ví dụ:

- Người dùng có quyền truy cập.
- Dữ liệu đầu vào hợp lệ.
- Thị trường đang mở cửa.

---

## Validation Rules

Kiểm tra dữ liệu trong quá trình thực thi.

Ví dụ:

- Mã cổ phiếu hợp lệ.
- Báo cáo tài chính tồn tại.
- Dữ liệu không bị thiếu.

---

## Completion Rules

Điều kiện để Runtime được xem là hoàn thành.

Ví dụ:

- Đã phân tích đủ dữ liệu.
- Đã tạo báo cáo.
- Đã lưu kết quả.

---

## Failure Rules

Điều kiện khiến Runtime thất bại.

Ví dụ:

- Không tìm thấy dữ liệu bắt buộc.
- Vi phạm ràng buộc nghiệp vụ.
- Thiếu quyền truy cập.

---

# Rule Evaluation

Business Rules được đánh giá tại các thời điểm sau.

- Trước khi Runtime bắt đầu.
- Sau khi Activity hoàn thành.
- Trước khi Runtime kết thúc.

---

# Rule Result

Mỗi Rule trả về một kết quả.

```
Pass
```

hoặc

```
Fail
```

Execution Engine sử dụng kết quả này để quyết định bước tiếp theo.

---

# Rule Ownership

Business Rules thuộc về Definition Models.

Ví dụ:

```
Workflow Definition

└── Business Rules
```

```
Stage Definition

└── Business Rules
```

```
Task Definition

└── Business Rules
```

Activity Definition cũng có thể định nghĩa Rule nếu cần.

---

# Integration

Business Rules làm việc với các thành phần sau.

## Definition Models

Cung cấp các Rule cần đánh giá.

---

## Context Model

Cung cấp dữ liệu đầu vào để đánh giá Rule.

---

## Orchestrator

Đọc kết quả đánh giá Rule để quyết định Runtime tiếp theo.

---

# Boundaries

Business Rules chịu trách nhiệm:

- Đánh giá điều kiện nghiệp vụ.
- Trả kết quả đánh giá.

Business Rules không chịu trách nhiệm:

- Điều phối Runtime.
- Chuyển trạng thái.
- Thực thi Activity.
- Quản lý Context.
- Thực hiện Retry hoặc Timeout.

---

# Design Principles

## Business First

Mọi Rule phản ánh yêu cầu nghiệp vụ.

---

## Definition Driven

Rule được định nghĩa trong Definition Models.

---

## Deterministic

Cùng một đầu vào phải cho cùng một kết quả.

---

## Independent

Rule không phụ thuộc vào Runtime Engine.

---

## Extensible

Có thể bổ sung Rule mới mà không thay đổi Execution Engine.

---

# Example

Workflow:

```
Analyze Company
```

Entry Rule:

- Company ID phải tồn tại.

Validation Rule:

- Báo cáo tài chính phải đầy đủ.

Completion Rule:

- Báo cáo phân tích đã được tạo.

Failure Rule:

- Không có dữ liệu tài chính trong ba năm gần nhất.

---

# Future Enhancements

Các khả năng sau không thuộc phạm vi phiên bản hiện tại:

- Rule Engine
- Expression Language
- Dynamic Rules
- Rule Versioning
- Rule Composition

---

# Related Documents

- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md
- context-model.md
- orchestration-model.md
- execution-model.md
- state-machine.md
