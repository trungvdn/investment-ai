# Architecture Overview

# Giới thiệu

Investment AI được xây dựng theo kiến trúc phân lớp (Layered Architecture), trong đó mỗi Layer đảm nhận một nhóm trách nhiệm riêng biệt.

Việc phân tách rõ ràng giữa các Layer giúp hệ thống dễ bảo trì, dễ mở rộng và cho phép thay thế từng thành phần mà không ảnh hưởng đến toàn bộ hệ thống.

Trong phiên bản MVP, kiến trúc được thiết kế để giải quyết bài toán phân tích doanh nghiệp và sinh báo cáo đầu tư.

---

# Kiến trúc tổng thể

```text
                        Client
                           │
                           ▼
                  Workflow Request
                           │
                           ▼
                  Execution Layer
                           │
                           ▼
                  Capability Layer
                           │
                           ▼
                Infrastructure Layer
                           │
      ┌────────────────────┼────────────────────┐
      │                    │                    │
   External APIs      Databases          LLM Providers
```

Ba Layer tạo thành lõi của hệ thống và phối hợp với nhau để thực hiện một Workflow.

---

# Kiến trúc phân lớp

## Execution Layer

Execution Layer chịu trách nhiệm điều phối việc thực thi Workflow.

Execution Layer trả lời câu hỏi:

> **Hệ thống cần thực hiện bước nào tiếp theo?**

Các thành phần chính:

- Workflow Runtime
- State Machine
- Orchestrator
- Activity Executor
- Execution Policy

Execution Layer không chứa Business Logic.

---

## Capability Layer

Capability Layer hiện thực toàn bộ Business Logic.

Capability Layer trả lời câu hỏi:

> **Nghiệp vụ được thực hiện như thế nào?**

Ví dụ:

- Phân tích báo cáo tài chính.
- Phân tích định giá.
- Phân tích rủi ro.
- Sinh báo cáo đầu tư.

Capability không quan tâm Workflow được điều phối ra sao.

---

## Infrastructure Layer

Infrastructure Layer cung cấp các dịch vụ kỹ thuật dùng chung.

Infrastructure Layer trả lời câu hỏi:

> **Làm thế nào để kết nối với tài nguyên bên ngoài?**

Ví dụ:

- Repository
- HTTP Client
- LLM Gateway
- Logger
- Cache
- Configuration

Infrastructure không chứa Business Logic.

Infrastructure không điều phối Workflow.

---

# Quan hệ giữa các Layer

```text
Execution
    │
    ▼
Capability
    │
    ▼
Infrastructure
```

Execution Layer sử dụng Capability để thực hiện nghiệp vụ.

Capability sử dụng Infrastructure để truy cập tài nguyên bên ngoài.

Infrastructure không phụ thuộc vào Business Logic.

---

# Luồng thực thi

Một Workflow được thực hiện theo trình tự sau.

```text
Workflow Request

↓

Workflow Definition

↓

Execution Runtime

↓

Activity

↓

Capability

↓

Infrastructure

↓

External Systems

↓

Capability

↓

Execution Runtime

↓

Workflow Result
```

---

# Quy tắc phụ thuộc

Kiến trúc tuân thủ các nguyên tắc sau.

## Execution không biết Implementation

Execution chỉ biết:

- Activity
- Handler Interface

Execution không biết:

- HTTP
- Database
- LLM

---

## Capability không điều phối Workflow

Capability chỉ thực hiện nghiệp vụ.

Capability không được:

- Chuyển State
- Điều phối Runtime
- Điều khiển Workflow

---

## Infrastructure không chứa Business Logic

Infrastructure chỉ:

- Gọi API
- Đọc Database
- Gọi LLM
- Ghi Log

Infrastructure không được:

- Phân tích dữ liệu
- Đưa ra kết luận đầu tư
- Điều phối Workflow

---

# Nguyên tắc thiết kế

## Separation of Concerns

Mỗi Layer chỉ đảm nhận một nhóm trách nhiệm.

---

## Single Responsibility

Mỗi thành phần chỉ có một trách nhiệm duy nhất.

---

## Dependency Inversion

Các Layer phụ thuộc vào abstraction thay vì implementation.

---

## Replaceable Components

Có thể thay thế:

- LLM
- Database
- API Provider
- Storage

mà không ảnh hưởng đến Business Logic.

---

## Extensibility

Kiến trúc cho phép mở rộng:

- Workflow mới
- Capability mới
- Infrastructure mới

mà không cần thay đổi các Layer còn lại.

---

# Kiến trúc MVP

Phiên bản đầu tiên của hệ thống chỉ bao gồm một Workflow:

```text
Analyze Company

↓

Load Data

↓

Analyze

↓

Generate Report
```

Mặc dù đơn giản, kiến trúc vẫn giữ nguyên nguyên tắc phân tách giữa Execution, Capability và Infrastructure để đảm bảo khả năng mở rộng trong các phiên bản tiếp theo.

---

# Tài liệu liên quan

- System Overview
- Capability Layer
- Execution Layer
- Infrastructure Layer

Architecture Overview là tài liệu mô tả cấu trúc tổng thể của hệ thống và là nền tảng cho toàn bộ các tài liệu kiến trúc ở các Layer phía dưới.
