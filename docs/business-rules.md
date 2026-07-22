# Business Rules

Version: 1.0

---

# 1. Purpose

Business Rules định nghĩa các quy tắc nghiệp vụ được sử dụng bởi mọi Capability trong Investment AI.

Business Rules là nguồn chân lý (Source of Truth) cho việc:

* Đánh giá dữ liệu.
* Sinh Evidence.
* Chấm điểm Analysis.
* Tạo Recommendation.

Business Rules không phụ thuộc:

* AI Provider
* Go
* Database
* Workflow

---

# 2. Rule Hierarchy

```text
Business Policy
        │
        ▼
Business Rule
        │
        ▼
Evaluation Rule
        │
        ▼
Evidence
```

---

# 3. Rule Categories

## Data Validation Rules

Kiểm tra tính hợp lệ của dữ liệu.

Ví dụ:

* Financial Statement phải có đủ 3 báo cáo.
* Volume không được âm.
* ROE phải nằm trong khoảng hợp lý.

---

## Evidence Generation Rules

Biến dữ liệu thành Evidence.

Ví dụ:

Nếu:

ROE > 20%

Trong ít nhất 5 năm

↓

Sinh Evidence:

"Khả năng sinh lời cao và ổn định."

---

## Signal Detection Rules

Ví dụ:

EMA20 cắt lên EMA50

↓

Golden Cross

---

## Analysis Rules

Ví dụ:

Nếu:

* Earnings Growth mạnh
* ROE cao
* Debt thấp

↓

Financial Score tăng.

---

## Decision Rules

Ví dụ:

Nếu:

* Financial Score > 80
* Technical Score > 70
* Risk thấp

↓

Recommendation = Buy

---

# 4. Rule Structure

Mỗi Rule gồm:

* Rule ID
* Name
* Purpose
* Inputs
* Conditions
* Outputs
* Priority
* Version

---

# 5. Rule Priority

Nếu nhiều Rule xung đột:

1. Safety Rules
2. Risk Rules
3. Financial Rules
4. Technical Rules
5. Reporting Rules

---

# 6. Rule Traceability

Mọi Rule phải truy vết được:

```text
Recommendation
        │
        ▼
Decision Rule
        │
        ▼
Analysis Rule
        │
        ▼
Evidence Rule
        │
        ▼
Raw Data
```

---

# 7. Rule Versioning

Rule phải có Version.

Ví dụ:

* BR-001 v1.0
* BR-001 v1.1

Thay đổi Rule không được làm mất khả năng tái hiện kết quả cũ.

---

# 8. Rule Governance

Mỗi Rule phải có:

* Owner
* Reviewer
* Effective Date
* Status

Status:

* Draft
* Active
* Deprecated
* Archived

---

# 9. Rule Design Principles

* Rule phải rõ ràng.
* Rule phải kiểm thử được.
* Rule không phụ thuộc implementation.
* Rule có thể thay thế hoặc mở rộng mà không ảnh hưởng Workflow.
* AI chỉ thực thi hoặc hỗ trợ diễn giải Rule, không tự ý thay đổi Rule.
