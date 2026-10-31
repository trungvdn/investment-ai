# Bộ khuôn mẫu Wyckoff — 9 cổ phiếu

Bộ mẫu gồm các mã: BCM, CTD, GEX, HDG, LPB, MSB, NLG, NVL, SZC.

## Cách điền

1. Bắt đầu bằng TICKER/report.md. Điền ngày phân tích và phạm vi dữ liệu trước khi nhận định.
2. Sao chép khối E001 cho mỗi sự kiện, đặt event_id duy nhất và liên kết structure_id. Không bắt buộc có đủ chuỗi sự kiện.
3. Ghi rõ cả nhãn đúng, nhãn sai và trường hợp chưa rõ. Nhãn sai phải giữ proposed_label, thêm final_label và correction_reason. Dùng no_event nếu không phải sự kiện Wyckoff.
4. Hoàn thiện lập luận, bằng chứng phản đối, điều kiện xác nhận/bác bỏ và kịch bản.
5. Duyệt toàn mẫu; chỉ đổi approved và eligible_for_ai=true khi đã kiểm tra dữ liệu, nhãn và lập luận.
6. Đồng bộ nội dung sang annotation.json. Nếu chỉ điền Markdown, có thể chuyển sang JSON sau; không duy trì hai kết luận khác nhau.

## Quy ước dữ liệu

- null = chưa điền/chưa xác định; [] = chưa có bản ghi, không có nghĩa đã xác nhận không có sự kiện.
- events để trống có chủ ý. Sao chép event-template.json vào events cho từng sự kiện thực tế và điền mọi trường cần thiết.
- Ngày dùng YYYY-MM-DD. Ngày sự kiện khác ngày đủ bằng chứng xác nhận.
- event_status là trạng thái thị trường: candidate, confirmed, invalidated, uncertain.
- label_verdict là kết quả con người duyệt nhãn: unreviewed, correct, incorrect, uncertain. Một sự kiện được nhận diện đúng vẫn có thể thất bại về sau.
- review.status: draft, in_review, needs_changes, approved. Đây là giá trị khởi tạo, không phải xác nhận của người dùng.
- structure: accumulation, distribution, reaccumulation, redistribution, unclear; phase: A/B/C/D/E/unclear. Không ép nhãn khi thiếu dữ liệu.
- Nhãn có thể dùng: PS, SC, AR, ST, Spring, Test, SOS, LPS, PSY, BC, UT, UTAD, SOW, LPSY, no_event. Gắn với cấu trúc và thứ tự thực tế; không dùng danh sách như chuỗi bắt buộc.
- Mỗi trading_ranges có structure_id, start_date, end_date, support_zone, resistance_zone, evidence. Vùng giá dùng lower/upper; không tự điền số chưa kiểm chứng.
- Mỗi scenarios có scenario_id, description, confirmation_condition, invalidation_condition, evidence_event_ids.
- Bằng chứng cần ngày/đoạn dữ liệu cụ thể. Nếu so volume với trung bình, ghi cửa sổ tính; nếu mô tả biên độ rộng, ghi mốc so sánh.
- Giá trị từ dữ liệu sau analysis_as_of không được dùng trong phần nhận định. future_outcome_excluded_from_input phải bị loại khỏi đầu vào AI khi đánh giá hồi cứu.
- Biểu đồ có nhãn là tham chiếu gốc chưa duyệt, không tự coi là ground truth. Khi phân tích hồi cứu, không cho AI xem phần biểu đồ sau ngày phân tích.
- data.ohlcv_path và data.chart_path được tính tương đối từ thư mục chứa annotation.json. Các liên kết trong report.md cũng trỏ tới tệp cùng thư mục. Trước khi duyệt, ghi commit nguồn để cố định phiên bản dữ liệu.

## Phạm vi bàn giao

Mỗi thư mục mã chứa đủ 4 tệp: report.md, annotation.json, candles_TICKER_1D.csv và TICKER_nhan.png. OHLCV và ảnh được chuyển từ data/sample vào cùng thư mục mẫu, giữ nguyên nội dung. Không gán nhãn hoặc kết luận thay người duyệt.
