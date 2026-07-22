# RFC-004 — Market Intelligence Business Rules

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team  
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Business Rules định nghĩa các nguyên tắc nghiệp vụ (Business Policies), ràng buộc (Constraints), bất biến (Invariants) và quy tắc quản trị (Governance Rules) của Market Intelligence Capability.

Đây là tài liệu quy định **hệ thống phải phân tích như thế nào**, chứ không quy định **phân tích bằng công thức nào**.

Các thuật toán, chỉ báo và quy tắc phân tích chi tiết được định nghĩa trong từng **Analysis Module**.

---

# 2. Scope

Business Rules áp dụng cho toàn bộ Market Intelligence Capability.

Bao gồm:

- Market Trend Analysis
- Liquidity Analysis
- Breadth Analysis
- Capital Flow Analysis
- Macro Analysis
- Market Cycle Analysis
- Risk Environment Analysis
- Context Synthesizer

Không áp dụng cho:

- Financial Analysis
- Company Intelligence
- Technical Analysis
- Portfolio Management

---

# 3. Rule Hierarchy

```
Business Policy
        │
        ▼
Business Constraint
        │
        ▼
Business Invariant
        │
        ▼
Analysis Module Rules
```

Trong đó:

- Business Policy là nguyên tắc cấp cao.
- Business Constraint là giới hạn nghiệp vụ.
- Business Invariant là điều kiện luôn đúng.
- Analysis Module Rules được định nghĩa trong từng Module.

---

# 4. Business Policies

## BP-001 — Multi-Dimensional Analysis

Mọi đánh giá thị trường phải được thực hiện từ nhiều góc nhìn.

Ví dụ:

- Trend
- Liquidity
- Breadth
- Capital Flow
- Macro
- Market Cycle
- Risk Environment

Không được kết luận chỉ từ một góc nhìn.

---

## BP-002 — Evidence-Based Decision

Mọi kết luận đều phải dựa trên Evidence.

Không được đưa ra kết luận dựa trên giả định hoặc suy đoán.

---

## BP-003 — Indicator Interpretation

Indicator chỉ là dữ liệu quan sát.

Indicator không mang ý nghĩa đầu tư nếu chưa được diễn giải bằng Business Rules.

---

## BP-004 — Context Before Decision

Market Intelligence chỉ tạo Market Context.

Không tạo:

- Buy Signal
- Sell Signal
- Investment Recommendation
- Portfolio Decision

---

## BP-005 — Explainability

Mọi kết luận phải có khả năng giải thích.

Người sử dụng phải truy vết được:

Summary

↓

Market Context

↓

Analysis Result

↓

Signal

↓

Evidence

↓

Indicator

---

## BP-006 — Consistency

Cùng dữ liệu đầu vào phải tạo cùng kết quả.

Không phụ thuộc:

- AI Model
- Programming Language
- Data Provider
- Infrastructure

---

## BP-007 — Independence

Các Analysis Module hoạt động độc lập.

Một Module không được sửa kết quả của Module khác.

---

## BP-008 — Separation of Concerns

Business Rules không chứa:

- Indicator Formula
- Technical Indicator Parameters
- Machine Learning Model
- Prompt
- Source Code

Những nội dung này thuộc Implementation hoặc Analysis Module.

---

# 5. Business Constraints

## BC-001

Market Context phải bao gồm tất cả Analysis Modules bắt buộc.

---

## BC-002

Market Context không được thiếu Evidence.

---

## BC-003

Signal không được tồn tại nếu không có Evidence.

---

## BC-004

Evidence không được tồn tại nếu không có Indicator.

---

## BC-005

Mỗi Analysis Result chỉ thuộc về một Analysis Module.

---

## BC-006

Market Analysis Result chỉ được sinh sau khi tất cả Analysis Modules hoàn thành hoặc được đánh dấu là không khả dụng.

---

# 6. Business Invariants

Các điều kiện dưới đây luôn đúng.

---

## BI-001

Indicator không chứa Recommendation.

---

## BI-002

Evidence luôn tham chiếu đến ít nhất một Indicator.

---

## BI-003

Signal luôn tham chiếu đến ít nhất một Evidence.

---

## BI-004

Market Context luôn được tạo từ Analysis Results.

---

## BI-005

Market Analysis Result luôn chứa đúng một Market Context.

---

## BI-006

Knowledge Contract phải có Version.

---

## BI-007

Business Rule phải có Identifier.

---

## BI-008

Analysis Result luôn có Confidence.

---

# 7. Decision Principles

## DP-001

Nếu Evidence chưa đủ mạnh:

Không đưa ra kết luận.

---

## DP-002

Nếu nhiều Evidence mâu thuẫn:

Không loại bỏ Evidence.

Đánh dấu Conflict.

---

## DP-003

Nếu Module không đủ dữ liệu:

Không suy diễn.

Trả về trạng thái:

Insufficient Data

---

## DP-004

Nếu Confidence thấp:

Được phép tạo Summary.

Không được khẳng định chắc chắn.

---

# 8. Conflict Resolution Policies

## CR-001

Conflict không phải Error.

Conflict là trạng thái bình thường của thị trường.

---

## CR-002

Conflict phải được ghi nhận.

Không được bỏ qua.

---

## CR-003

Nếu hai Module đưa ra kết luận trái ngược:

Context Synthesizer chịu trách nhiệm tổng hợp.

Không Module nào được quyền ghi đè Module khác.

---

## CR-004

Conflict phải làm giảm Confidence của Market Context.

Không được làm mất Evidence.

---

# 9. Governance Rules

## GR-001

Mỗi Business Rule phải có:

- Rule Identifier
- Name
- Description
- Owner
- Version
- Effective Date

---

## GR-002

Business Rule không được sửa trực tiếp.

Mọi thay đổi phải tạo Version mới.

---

## GR-003

Business Rule phải được Review trước khi phát hành.

---

## GR-004

Business Rule phải truy vết được tới tài liệu nguồn hoặc giả định nghiệp vụ khi cần.

---

# 10. Rule Lifecycle

```
Draft
    │
    ▼
Review
    │
    ▼
Approved
    │
    ▼
Active
    │
    ▼
Deprecated
    │
    ▼
Archived
```

Rule đã Active không được chỉnh sửa.

Chỉ được:

- Deprecate
- Replace
- Version Up

---

# 11. Rule Ownership

| Rule Type | Owner |
|-----------|-------|
| Business Policies | Domain Architect |
| Business Constraints | Domain Architect |
| Business Invariants | Domain Architect |
| Analysis Rules | Analysis Module |
| Context Rules | Context Synthesizer |

---

# 12. Extension Rules

Được phép mở rộng:

- Business Policy mới
- Constraint mới
- Invariant mới
- Conflict Policy mới

Không được:

- Thay đổi ngữ nghĩa của Rule hiện tại.
- Xóa Rule đang Active.
- Gộp nhiều Rule có mục đích khác nhau thành một Rule.

---

# 13. Related Documents

- RFC-001 — Market Intelligence Specification
- RFC-002 — Market Intelligence Information Model
- RFC-003 — Market Intelligence Knowledge Contracts
- Analysis Module Specifications
