# NLG — Hồ sơ phân tích Wyckoff

> KHUÔN MẪU CHƯA ĐIỀN. Không phải kết luận phân tích hay mẫu đã duyệt.

## 1. Thông tin và dữ liệu

- Mã: NLG
- Khung thời gian: 1D
- OHLCV nguồn: [candles_NLG_1D.csv](candles_NLG_1D.csv)
- Biểu đồ tham chiếu: [NLG_nhan.png](NLG_nhan.png)
- Ngày phân tích (analysis_as_of): [YYYY-MM-DD]
- Khoảng dữ liệu thực sự sử dụng: […] → […]
- Phiên bản/commit dữ liệu: […]
- Giá và khối lượng đã điều chỉnh? […]
- Kiểm tra thiếu/trùng phiên, đơn vị giá và volume: […]
- Người phân tích: […]

## 2. Kết luận nhanh

- Cấu trúc: […]
- Pha: […]
- Đang hình thành / Đã xác nhận / Thất bại / Chưa rõ: […]
- Mức tin cậy thấp / trung bình / cao và lý do: […]
- Luận điểm chính: […]
- Bằng chứng còn thiếu: […]

## 3. Bối cảnh trước cấu trúc

- Xu hướng trước đó, thời gian và mức biến động: […]
- Dấu hiệu thay đổi hành vi cung–cầu: […]
- Bối cảnh thị trường/ngành nếu có (nguồn và ngày): […]
- Giả thuyết ưu tiên và lý do: […]
- Giả thuyết cạnh tranh và bằng chứng: […]

## 4. Trading Range

| Thành phần | Ngày/vùng giá | Bằng chứng |
|---|---|---|
| Bắt đầu vùng | […] | […] |
| Kết thúc vùng hoặc còn tiếp diễn | […] | […] |
| Hỗ trợ | […] | […] |
| Kháng cự | […] | […] |
| Hành vi bên trong vùng | […] | […] |

## 5. Nhật ký nhãn sự kiện

Sao chép khối dưới đây cho MỖI sự kiện ứng viên. Giữ cả nhãn bị bác bỏ; không xóa lịch sử sửa nhãn. Mỗi Trading Range có structure_id riêng.

### Sự kiện E001

- structure_id: […]
- Nhãn đề xuất ban đầu: […]
- Ngày bắt đầu / kết thúc: […] / […]
- Ngày có đủ bằng chứng để xác nhận (nếu có): […]
- Nguồn đề xuất: người / AI / biểu đồ tham chiếu: […]
- Bối cảnh và sự kiện liên quan trước đó: […]
- Bằng chứng giá (ngày, OHLC, biên độ, vị trí đóng cửa, vùng giá): […]
- Bằng chứng volume (giá trị, mốc so sánh và cửa sổ tính): […]
- Phản ứng sau sự kiện, chỉ đến ngày phân tích: […]
- Bằng chứng phản đối: […]
- Điều kiện xác nhận, có thể kiểm tra được: […]
- Điều kiện bác bỏ, có thể kiểm tra được: […]
- Trạng thái sự kiện: candidate / confirmed / invalidated / uncertain: […]
- Kết quả duyệt nhãn: unreviewed / correct / incorrect / uncertain: […]
- Nhãn cuối cùng sau duyệt (có thể là no_event): […]
- Lý do giữ hoặc sửa nhãn: […]
- Người duyệt / ngày duyệt: […] / […]

## 6. Cung–cầu và pha

| Khía cạnh | Quan sát có ngày tham chiếu | Diễn giải | Bằng chứng phản đối |
|---|---|---|---|
| Nhịp tăng so với nhịp giảm | […] | […] | […] |
| Nỗ lực so với kết quả | […] | […] | […] |
| Kiểm tra hỗ trợ | […] | […] | […] |
| Kiểm tra kháng cự | […] | […] | […] |

- Pha ưu tiên và các event_id hỗ trợ: […]
- Vì sao chưa chọn pha/giả thuyết khác: […]
- Điều kiện chuyển pha: […]
- Điều kiện bác bỏ toàn bộ luận điểm: […]

## 7. Kịch bản

| Kịch bản | Điều kiện xác nhận | Điều kiện bác bỏ | Tín hiệu cần theo dõi |
|---|---|---|---|
| Phát triển tích cực | […] | […] | […] |
| Tiếp tục đi ngang | […] | […] | […] |
| Suy yếu/thất bại | […] | […] | […] |

## 8. Duyệt toàn bộ mẫu

- Trạng thái: draft / in_review / needs_changes / approved: […]
- Người duyệt / ngày duyệt: […] / […]
- Các nhãn đúng đã kiểm tra: […]
- Các nhãn sai và cách sửa: […]
- Những điểm chưa thể kết luận: […]
- Yêu cầu bổ sung: […]
- Đủ điều kiện dùng làm mẫu AI? có / không, lý do: […]

## 9. Kết quả tương lai — tách khỏi đầu vào AI

Chỉ điền sau này để đánh giá; không đưa vào prompt tại analysis_as_of.

- Ngày đánh giá: […]
- Diễn biến thực tế sau ngày phân tích: […]
- Kịch bản được xác nhận/bác bỏ: […]
- Sai ở dữ liệu, nhãn, lập luận hay tín hiệu chưa đủ: […]
- Bài học: […]
