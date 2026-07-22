# Data Acquisition Business Rules

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team

---

# 1. Purpose

Tài liệu này định nghĩa các Business Rules của Data Acquisition Capability.

Business Rules quy định cách Data Acquisition thu thập, kiểm tra, chuẩn hóa và phân phối dữ liệu nhằm đảm bảo toàn bộ hệ thống sử dụng dữ liệu nhất quán, đáng tin cậy và có thể truy vết.

---

# 2. Business Policies

## Single Source of Truth

Mỗi Data Object chỉ được publish một phiên bản hợp lệ tại một thời điểm.

---

## Canonical Data Model

Tất cả dữ liệu phải được chuyển đổi sang Canonical Data Model trước khi publish.

Business Capability không được phụ thuộc trực tiếp vào định dạng dữ liệu của Data Provider.

---

## Provider Independence

Business Capability không được biết dữ liệu đến từ Provider nào.

Việc lựa chọn Provider là trách nhiệm của Data Acquisition.

---

## Metadata Required

Mọi dữ liệu được publish phải có Metadata đầy đủ.

Ví dụ:

- Source
- Timestamp
- Version

---

## Quality First

Chỉ dữ liệu đạt yêu cầu chất lượng mới được publish.

---

# 3. Validation Rules

Mọi dữ liệu phải được kiểm tra trước khi sử dụng.

Bao gồm:

- Required Fields
- Data Type
- Format
- Business Constraints
- Duplicate Detection

Nếu Validation thất bại, dữ liệu không được publish.

---

# 4. Normalization Rules

Mọi dữ liệu phải được chuẩn hóa trước khi publish.

Bao gồm:

- Naming Convention
- Timezone
- Currency
- Symbol
- Date Format
- Number Format

Business Capability luôn làm việc với dữ liệu đã chuẩn hóa.

---

# 5. Publication Rules

Một Data Object chỉ được publish khi:

- Validation thành công.
- Normalization hoàn tất.
- Metadata đầy đủ.
- Quality Assessment đạt yêu cầu.

Nếu một trong các điều kiện không đạt, dữ liệu không được publish.

---

# 6. Versioning Rules

Mọi Knowledge Contract đều phải có Version.

Version mới không được làm ảnh hưởng đến Consumer hiện tại nếu không có Breaking Change.

---

# 7. Consistency Rules

Các dữ liệu liên quan phải nhất quán với nhau.

Ví dụ:

- Company Data phải nhất quán với Financial Data.
- Market Data phải đồng bộ với Trading Calendar.
- Corporate Actions không được mâu thuẫn với Company Information.

---

# 8. Freshness Rules

Capability phải luôn ưu tiên dữ liệu mới nhất.

Nếu dữ liệu đã lỗi thời:

- Không được coi là dữ liệu hiện hành.
- Metadata phải phản ánh thời điểm cập nhật cuối cùng.

---

# 9. Traceability Rules

Mọi dữ liệu phải có khả năng truy vết.

Có thể xác định:

- Data Source
- Provider
- Version
- Publish Time

---

# 10. Error Handling Rules

Nếu xảy ra lỗi trong quá trình thu thập:

- Không publish dữ liệu chưa hoàn chỉnh.
- Không ghi đè dữ liệu hợp lệ hiện có.
- Lưu lại thông tin lỗi để phục vụ điều tra.

---

# 11. Governance Rules

Data Acquisition là Capability duy nhất chịu trách nhiệm:

- Thu thập dữ liệu.
- Chuẩn hóa dữ liệu.
- Kiểm tra dữ liệu.
- Publish dữ liệu.

Các Capability khác không được:

- Thu thập trực tiếp từ Data Provider.
- Chuẩn hóa lại dữ liệu.
- Tự sửa đổi dữ liệu đã publish.

---

# 12. Decision Principles

Khi có nhiều nguồn dữ liệu cho cùng một Business Object:

- Ưu tiên nguồn dữ liệu có chất lượng cao hơn.
- Ưu tiên nguồn dữ liệu chính thức.
- Giữ tính nhất quán trong cùng một phiên bản dữ liệu.

Chiến lược lựa chọn Provider được cấu hình bên ngoài Business Rules.

---

# 13. Conflict Resolution

Khi dữ liệu từ nhiều Provider không nhất quán:

- Không tự động hợp nhất dữ liệu.
- Áp dụng chính sách lựa chọn Provider đã cấu hình.
- Ghi nhận nguồn dữ liệu được sử dụng.
- Lưu thông tin phục vụ kiểm toán nếu cần.

---

# 14. Invariants

Các nguyên tắc sau luôn phải đúng:

- Published Data luôn đã được Validation.
- Published Data luôn đã được Normalization.
- Published Data luôn có Metadata.
- Published Data luôn có Source.
- Published Data luôn có Version.
- Business Capability không phụ thuộc trực tiếp vào Data Provider.
- Canonical Data Model là định dạng dữ liệu duy nhất được publish.

---

# 15. Related Documents

- README.md
- specification.md
- information-model.md
- knowledge-contracts.md
