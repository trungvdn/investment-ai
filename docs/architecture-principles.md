# Architecture Principles

Version: 1.0

---

# 1. Purpose

Tài liệu này định nghĩa các nguyên tắc kiến trúc (Architecture Principles) cho toàn bộ hệ thống Investment AI.

Mọi quyết định về thiết kế, phát triển và mở rộng hệ thống phải tuân theo các nguyên tắc này.

Architecture Principles có mức ưu tiên cao hơn mọi tài liệu thiết kế chi tiết.

---

# 2. Domain First

Business Domain là trung tâm của hệ thống.

Kiến trúc phải phản ánh quy trình đầu tư thực tế thay vì phản ánh công nghệ sử dụng.

Không thiết kế hệ thống xoay quanh:

* LLM
* Agent
* Framework
* Prompt

Mà phải xoay quanh:

* Investment Process
* Business Capability
* Business Rule

---

# 3. Capability Before Agent

Capability là năng lực nghiệp vụ.

Agent chỉ là một cách triển khai capability.

Một capability có thể được hiện thực bằng:

* Rule Engine
* AI Model
* Machine Learning
* Agent
* Hybrid

Capability không được phụ thuộc vào Agent.

---

# 4. Workflow Driven

Workflow quyết định cách hệ thống hoạt động.

Workflow phải mô tả chính xác quy trình đầu tư.

Workflow không được phụ thuộc vào LLM.

Workflow chỉ điều phối các capability.

---

# 5. Data Is The Source of Truth

Mọi kết luận phải xuất phát từ dữ liệu.

Prompt không được tạo ra dữ liệu mới.

LLM không được phép suy diễn nếu thiếu bằng chứng.

Nếu dữ liệu không đủ phải trả về:

* Unknown
* Insufficient Evidence

Thay vì đoán.

---

# 6. Evidence Before Conclusion

Không được đưa ra Recommendation trước khi có Evidence.

Mỗi Recommendation phải truy vết được:

Recommendation

↓

Investment Thesis

↓

Evidence

↓

Raw Data

Hệ thống phải luôn có khả năng giải thích ngược.

---

# 7. Explainability By Default

Mọi output đều phải trả lời được:

* Vì sao?
* Dựa trên dữ liệu nào?
* Độ tin cậy bao nhiêu?
* Có rủi ro gì?

Nếu không giải thích được thì không được coi là output hợp lệ.

---

# 8. Structured Output First

Mọi capability phải trả về dữ liệu có cấu trúc.

Ví dụ:

* JSON
* Typed Object
* Domain Model

Không truyền văn bản tự do giữa các capability.

Natural Language chỉ được tạo ở lớp Reporting.

---

# 9. Stateless Capability

Capability không lưu trạng thái lâu dài.

Input quyết định Output.

Capability có thể chạy độc lập.

Điều này giúp:

* Test dễ dàng
* Retry dễ dàng
* Scale dễ dàng

---

# 10. Stateful Workflow

Workflow chịu trách nhiệm quản lý trạng thái.

Ví dụ:

* Execution State
* Context
* Dependency
* Retry
* Progress

Capability không cần biết workflow đang ở bước nào.

---

# 11. Loose Coupling

Các capability phải giao tiếp thông qua Domain Model.

Không gọi trực tiếp vào implementation của nhau.

Chỉ phụ thuộc vào:

* Input Contract
* Output Contract

Điều này cho phép thay thế implementation mà không ảnh hưởng toàn hệ thống.

---

# 12. Single Responsibility

Một capability chỉ có một trách nhiệm.

Ví dụ:

Technical Analysis

Không thực hiện:

* Valuation
* Macro Analysis
* Report Generation

Mỗi capability chỉ giải quyết đúng bài toán của mình.

---

# 13. AI Is Replaceable

LLM là một thành phần có thể thay thế.

Kiến trúc không được phụ thuộc vào:

* Prompt cụ thể
* Model cụ thể
* Nhà cung cấp cụ thể

Thay đổi AI Provider không làm thay đổi Business Logic.

---

# 14. Progressive Intelligence

Hệ thống phải phát triển theo từng lớp.

Data

↓

Information

↓

Analysis

↓

Investment Thesis

↓

Recommendation

↓

Report

Không bỏ qua bất kỳ bước nào.

---

# 15. Reusable Components

Mọi capability phải có khả năng tái sử dụng.

Ví dụ:

Macro Analysis

có thể được sử dụng trong:

* Stock Analysis
* Portfolio Analysis
* Market Dashboard
* Risk Assessment

Không thiết kế capability chỉ phục vụ một workflow.

---

# 16. Observability

Mọi workflow đều phải có khả năng theo dõi.

Bao gồm:

* Execution Time
* Input
* Output
* Token Usage
* Cost
* Error
* Retry

Điều này giúp đánh giá chất lượng và tối ưu hệ thống.

---

# 17. Incremental Evolution

Ưu tiên xây dựng MVP trước.

Chỉ bổ sung:

* Agent Collaboration
* Memory
* Reflection
* RAG
* Planning

khi MVP đã chứng minh được giá trị.

Không xây dựng kiến trúc quá phức tạp cho những nhu cầu chưa tồn tại.

---

# 18. Decision Checklist

Trước mỗi quyết định kiến trúc, cần trả lời:

1. Có phản ánh đúng Business Domain không?
2. Có làm capability phụ thuộc vào AI không?
3. Có tăng khả năng tái sử dụng không?
4. Có dễ kiểm thử không?
5. Có thể giải thích kết quả không?
6. Có thể thay thế implementation không?
7. Có làm MVP phức tạp hơn không?

Nếu bất kỳ câu trả lời nào là "Không", cần xem xét lại thiết kế.

---

# 19. Architecture Hierarchy

Mọi thành phần trong hệ thống phải tuân theo thứ tự phụ thuộc sau:

Business Domain

↓

Business Capability

↓

Workflow

↓

Domain Model

↓

Capability Implementation

↓

Infrastructure

↓

AI Provider

Các lớp phía trên không được phụ thuộc vào các lớp phía dưới.

Đặc biệt:

* Business Domain không phụ thuộc AI.
* Workflow không phụ thuộc Model.
* Capability không phụ thuộc Prompt.
* Infrastructure không chứa Business Logic.

Đây là nguyên tắc quan trọng nhất của toàn bộ kiến trúc.
