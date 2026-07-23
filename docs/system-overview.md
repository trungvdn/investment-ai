# System Overview

# Giới thiệu

Investment AI là một hệ thống hỗ trợ phân tích và đánh giá doanh nghiệp bằng cách tự động hóa quy trình nghiên cứu đầu tư.

Hệ thống được thiết kế theo kiến trúc Workflow Engine, trong đó quy trình phân tích được mô hình hóa thành các Workflow có thể định nghĩa, thực thi và mở rộng.

Trong phiên bản MVP, hệ thống tập trung giải quyết một bài toán duy nhất:

> **Đầu vào:** Mã cổ phiếu

> **Đầu ra:** Báo cáo đánh giá doanh nghiệp

---

# Mục tiêu

Hệ thống được xây dựng nhằm các mục tiêu sau:

- Chuẩn hóa quy trình phân tích đầu tư.
- Tách biệt Business Logic khỏi Workflow Engine.
- Dễ dàng mở rộng thêm Workflow mới.
- Hỗ trợ nhiều nguồn dữ liệu và nhiều mô hình AI.
- Dễ dàng bảo trì và thay thế các thành phần kỹ thuật.

---

# Kiến trúc tổng thể

Hệ thống được chia thành ba Layer chính.

```text
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

Mỗi Layer đảm nhận một trách nhiệm riêng và giao tiếp thông qua các interface rõ ràng.

---

# Các Layer

## Execution Layer

Execution Layer chịu trách nhiệm điều phối quá trình thực thi Workflow.

Các nhiệm vụ chính:

- Quản lý Runtime.
- Điều phối Workflow.
- Thực thi Activity.
- Quản lý State.
- Quản lý Context.
- Điều khiển Execution Flow.

Execution Layer không chứa Business Logic và không biết cách giao tiếp với các hệ thống bên ngoài.

---

## Capability Layer

Capability Layer chứa toàn bộ Business Logic của hệ thống.

Mỗi Capability đại diện cho một năng lực nghiệp vụ độc lập.

Ví dụ:

- Phân tích báo cáo tài chính.
- Định giá doanh nghiệp.
- Phân tích rủi ro.
- Sinh báo cáo đầu tư.

Capability Layer tập trung vào việc giải quyết bài toán nghiệp vụ và không quan tâm đến cách Workflow được điều phối.

---

## Infrastructure Layer

Infrastructure Layer cung cấp các dịch vụ kỹ thuật cho toàn bộ hệ thống.

Các thành phần điển hình:

- Repository
- HTTP Client
- LLM Gateway
- Logger
- Cache
- Object Storage
- Message Queue
- Configuration

Infrastructure Layer chịu trách nhiệm kết nối hệ thống với các tài nguyên bên ngoài và trừu tượng hóa các chi tiết triển khai.

Infrastructure Layer không chứa Business Logic và không điều phối Workflow.

---

# Quan hệ giữa các Layer

```text
Execution Layer
        │
        ▼
Capability Layer
        │
        ▼
Infrastructure Layer
        │
        ▼
External Systems
```

Trong đó:

- Execution Layer quyết định **làm gì tiếp theo**.
- Capability Layer quyết định **nghiệp vụ được thực hiện như thế nào**.
- Infrastructure Layer quyết định **giao tiếp với tài nguyên bên ngoài bằng cách nào**.

---

# Luồng xử lý

Một yêu cầu phân tích cổ phiếu được thực hiện theo quy trình sau.

```text
Yêu cầu phân tích

↓

Workflow Definition

↓

Khởi tạo Runtime

↓

Execution Engine

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

Execution Engine

↓

Investment Report
```

Execution Engine không trực tiếp gọi API hay LLM.

Capability sẽ sử dụng các dịch vụ được cung cấp bởi Infrastructure để hoàn thành nghiệp vụ.

---

# Kiến trúc MVP

Phiên bản đầu tiên của hệ thống chỉ bao gồm một Workflow.

```text
Phân tích doanh nghiệp

↓

Thu thập dữ liệu

↓

Phân tích

↓

Sinh báo cáo
```

Các Workflow khác sẽ được bổ sung trong các phiên bản tiếp theo.

---

# Nguyên tắc thiết kế

Toàn bộ hệ thống được xây dựng dựa trên các nguyên tắc sau.

## Separation of Concerns

Mỗi Layer chỉ đảm nhận một nhóm trách nhiệm.

---

## Single Responsibility

Mỗi thành phần chỉ có một trách nhiệm duy nhất.

---

## Dependency Inversion

Business Logic không phụ thuộc vào hạ tầng.

Execution Engine không phụ thuộc vào Capability Implementation.

Infrastructure có thể được thay thế mà không ảnh hưởng đến các Layer còn lại.

---

## Extensibility

Có thể bổ sung:

- Workflow mới.
- Capability mới.
- Data Provider mới.
- LLM mới.
- Database mới.

mà không cần thay đổi kiến trúc hiện có.

---

# Roadmap

Kiến trúc được thiết kế để phát triển theo từng giai đoạn.

## Giai đoạn 1

- Một Workflow.
- Một báo cáo đầu tư.

## Giai đoạn 2

- Nhiều Workflow.
- Nhiều Capability.
- Nhiều nguồn dữ liệu.

## Giai đoạn 3

- AI Agent.
- Workflow cộng tác.
- Hệ thống nghiên cứu đầu tư tự động.

---

# Tài liệu liên quan

Để hiểu chi tiết hệ thống, nên đọc theo thứ tự sau:

1. System Overview
2. Capability Layer
3. Execution Layer
4. Infrastructure Layer

Ba Layer này tạo thành nền tảng của Workflow Engine:

- **Execution Layer** điều phối quy trình thực thi.
- **Capability Layer** hiện thực nghiệp vụ.
- **Infrastructure Layer** cung cấp các dịch vụ kỹ thuật và kết nối với thế giới bên ngoài.

Sự tách biệt này giúp hệ thống dễ bảo trì, dễ mở rộng và dễ thay thế từng thành phần mà không ảnh hưởng đến toàn bộ kiến trúc.
