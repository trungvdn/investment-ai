# HDG — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn, chờ người dùng duyệt. review.status=in_review; eligible_for_ai=false. Dựa trên OHLCV, chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_HDG_1D.csv](candles_HDG_1D.csv): 497 phiên từ 2024-10-31 đến 2026-10-30; cột event trống.
- [HDG_nhan.png](HDG_nhan.png): ảnh để người duyệt kiểm tra.
- source_commit: ee4c957bc6c22978be3461fec43b976c908e27c1.
- Close cuối 15 theo đơn vị CSV. Chưa xác minh lịch giao dịch, đơn vị và điều chỉnh giá.
- RV20=volume/trung bình 20 phiên liền trước, không tính phiên hiện tại.

## 2. Kết luận nhanh

**Cấu trúc hiện tại chưa rõ, pha UNKNOWN; ưu tiên bối cảnh markdown sau thất bại vùng tháng 7–9.** Vùng tháng 10 có khả năng dừng giảm nhưng chưa đủ bằng chứng xác nhận tích lũy hoặc pha D. Breakout 04/09 có xảy ra nhưng không giữ được; nhịp tăng 21/10 chưa tự tạo SOS/LPS mới.

## 3. Bối cảnh

- High 17/06=21.9; low 28/07=15.2 sau nhịp giảm mạnh tháng 7.\n- Đi ngang tháng 7–9 và breakout 04/09 không giữ được: close 02/10=15.05, 05/10=14.65 dưới đáy 15.2.\n- Nhịp hồi tháng 10 tới high 15.7 ngày 22/10; 30/10 close 15 dưới đáy vùng cũ.

Không suy ra danh tính/ý định tổ chức hoặc nguyên nhân doanh nghiệp từ OHLCV. Chưa thêm ngành/vĩ mô do thiếu dữ liệu trong mẫu.

## 4. Các vùng giao dịch

| structure_id | Thời gian | Hỗ trợ | Kháng cự | Trạng thái |
|---|---|---|---|---|
| HDG_TR_202607 | 27/07–05/10 | 15.2–15.6 | 17.0–17.2 | provisional; breakdown sau breakout thất bại |
| HDG_TR_202610 | Từ 05/10 | 14.15–14.3 | 15.6–15.7 | provisional; dừng giảm mới chưa xác nhận |

17.7 là đỉnh breakout 04/09, không phải biên đã biết trước breakout. Biên 15.6–15.7 vùng mới chỉ hình thành đến 22/10; không dùng ngược về 21/10 để giả vờ biết cạnh trên. end_date vùng cũ chỉ là mốc breakdown theo mẫu.

## 5. Nhật ký nhãn sự kiện

Ngoại trừ E005 đã được người dùng sửa và xác nhận, các sự kiện còn lại giữ label_verdict=unreviewed, final_label=null. invalidated là đề xuất của AI. Phân biệt **nhãn sai** với **sự kiện có xảy ra nhưng kịch bản thành công bị bác bỏ sau đó**, đặc biệt E004.

### E001 — SC: 2026-07-27 → 2026-07-28

- structure_id: HDG_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 27/07 low=close=15.5; 28/07 low 15.2, close 16.1 ở 90% biên độ.
- Bằng chứng volume: 27/07 volume 1610500 RV20 0.80; 28/07 1315800 RV20 0.64, không có volume climax.
- Phản ứng đến cutoff: Hồi tới high 17.2 ngày 10/08 rồi dao động.
- Bằng chứng phản đối/giới hạn: Chọn cụm stopping action ứng viên, không khẳng định Selling Climax chỉ vì đáy.
- Điều kiện xác nhận: Cần phản ứng hồi và kiểm tra sau đó; người duyệt có thể thay SC bằng no_event/stopping action.
- Điều kiện bác bỏ: Nếu thiếu bằng chứng climax thì bỏ SC; breakdown tháng 10 bác bỏ tích lũy thành công, không tự chứng minh nhãn dừng giảm lịch sử luôn sai.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E002 — AR: 2026-07-29 → 2026-08-10

- structure_id: HDG_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: High 10/08=17.2, close 17.05 sau cụm đáy 27–28/07.
- Bằng chứng volume: 10/08 volume 1462800 RV20 0.98.
- Phản ứng đến cutoff: Sau 10/08 quay về low 15.6 ngày 19/08.
- Bằng chứng phản đối/giới hạn: SC chưa chắc chắn; nhãn AR phụ thuộc cách xác định stopping action và phạm vi cấu trúc.
- Điều kiện xác nhận: Nhịp hồi sau dừng giảm tạo cạnh trên và có phản ứng giảm tiếp theo.
- Điều kiện bác bỏ: Nếu SC/cấu trúc gốc bị sửa thì xem lại AR.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E003 — ST: 2026-08-19

- structure_id: HDG_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 15.6, close 15.8; chưa phá SC low 15.2.
- Bằng chứng volume: 1566100; RV20 1.21, lớn hơn volume 28/07=1315800.
- Phản ứng đến cutoff: 21/08 close 16.35; 24/08 16.75.
- Bằng chứng phản đối/giới hạn: Không phải test volume giảm so phiên đáy; cần đánh giá thêm chất lượng cung.
- Điều kiện xác nhận: Giữ vùng hỗ trợ và có phản ứng hồi sau kiểm tra.
- Điều kiện bác bỏ: Nếu thiếu quan hệ kiểm tra với stopping action, bỏ ST thay vì ép chuỗi.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E004 — SOS: 2026-09-04

- structure_id: HDG_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 17.5, high 17.7 vượt biên 17.2 đã có trước.
- Bằng chứng volume: 3936500; RV20 3.72.
- Phản ứng đến cutoff: 07/09 close 17; 08/09 16.65 quay vào vùng; tháng 10 phá đáy.
- Bằng chứng phản đối/giới hạn: Breakout thực sự có xảy ra nhưng không duy trì; không dùng kết quả sau đó để xóa quan sát breakout ban đầu.
- Điều kiện xác nhận: Cần follow-through và giữ biên trên; điều kiện thành công không được duy trì tới cutoff.
- Điều kiện bác bỏ: Việc quay vào vùng rồi phá 15.2 bác bỏ giả thuyết SOS thành công/markup tiếp diễn, không tự đồng nghĩa nhãn breakout ban đầu sai.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E005 — Tiếp tục markdown sau khi Accumulation thất bại: 2026-10-02 → 2026-10-05

- structure_id: HDG_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 02/10 close 15.05 dưới 15.2; 05/10 low=close=14.65 tiếp diễn breakdown.
- Bằng chứng volume: 05/10 volume 1143400 RV20 1.33, tăng vừa phải.
- Phản ứng đến cutoff: 09/10 low 14.15; hồi 22/10 chỉ high 15.7, 30/10 close 15.
- Bằng chứng phản đối/giới hạn: Không có cấu trúc phân phối/tái phân phối được xác nhận để gán SOW; giữ diễn giải markdown tiếp diễn sau Accumulation thất bại.
- Điều kiện xác nhận: Breakout tích lũy 04/09 không giữ được; 02–05/10 đóng dưới đáy TR 15.2 và giá tiếp tục tạo đáy thấp hơn.
- Điều kiện bác bỏ: Lấy lại 15.2–15.6 và giữ qua kiểm tra làm suy yếu breakdown thành công.
- event_status: confirmed tại cutoff 30/10/2026; nhãn SOW ban đầu: incorrect.
- Nhãn cuối: markdown_continuation; người duyệt: trungvdn; ngày duyệt: 2026-10-31.
- Lý do sửa: Người dùng sửa E005: tiếp tục markdown sau khi Accumulation thất bại; không gán SOW của phân phối/tái phân phối.
- Đây là mô tả markdown tiếp diễn, không phải nhãn SOW kinh điển.

### E006 — SC: 2026-10-09

- structure_id: HDG_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 14.15, close 14.6 ở 81.8% biên độ.
- Bằng chứng volume: 1009600; RV20 1.32, không phải climax rõ.
- Phản ứng đến cutoff: Hồi đến high 15.7 ngày 22/10; chưa phá 14.15 tại cutoff.
- Bằng chứng phản đối/giới hạn: Dừng giảm ngắn hạn có thể chỉ là hồi kỹ thuật trong markdown.
- Điều kiện xác nhận: Cần AR và kiểm tra giữ đáy; không xác nhận SC chỉ vì là đáy thấp nhất.
- Điều kiện bác bỏ: Nếu tiếp tục markdown bền dưới 14.15, đánh giá lại stopping action hiện tại.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E007 — AR: 2026-10-12 → 2026-10-22

- structure_id: HDG_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 21/10 close=high=15.6; 22/10 high 15.7, close 15.45.
- Bằng chứng volume: 21/10 volume 2043600 RV20 2.75; 22/10 1151600 RV20 1.42.
- Phản ứng đến cutoff: 26–30/10 pullback về close 15.
- Bằng chứng phản đối/giới hạn: Đây là cạnh trên hình thành qua nhịp hồi; không biết sẵn 15.7 trước 22/10.
- Điều kiện xác nhận: Hồi sau dừng giảm và có pullback giúp xác định biên vùng mới.
- Điều kiện bác bỏ: Nếu không hình thành nền mà chỉ tiếp diễn markdown, AR của giả thuyết tích lũy cần xét lại.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E008 — ST: 2026-10-27 → 2026-10-30

- structure_id: HDG_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 28/10=14.95; close 30/10=15, vẫn cao hơn 14.15.
- Bằng chứng volume: 28/10 volume 377600; 30/10 751100 RV20 0.97.
- Phản ứng đến cutoff: Chưa có hồi sau 30/10 trong dữ liệu.
- Bằng chứng phản đối/giới hạn: Kiểm tra nông, chưa sát đáy; 30/10 đóng đáy phiên không cho thấy cầu đã xác nhận.
- Điều kiện xác nhận: Cần giữ đáy pullback và có phản ứng hồi, đánh giá ST hay pullback thông thường.
- Điều kiện bác bỏ: Nếu giảm tiếp thiếu cầu hỗ trợ thì không xác nhận ST.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E009 — Spring: 2026-10-09

- structure_id: HDG_TR_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 14.15, close 14.6 vẫn dưới hỗ trợ vùng cũ 15.2.
- Bằng chứng volume: Volume tăng không thay thế điều kiện hồi vào vùng.
- Phản ứng đến cutoff: 22/10 hồi lên 15.7 nhưng sau đó 30/10 close 15; không phải phục hồi nhanh ngay cụm xuyên đáy.
- Bằng chứng phản đối/giới hạn: Không ghép hồi rất muộn vào một Spring thành công mà bỏ qua thời gian giá chấp nhận dưới vùng.
- Điều kiện xác nhận: Spring của vùng cũ phải có xuyên và hồi lại vùng với bằng chứng phù hợp.
- Điều kiện bác bỏ: Đề xuất bác bỏ Spring thành công 09/10 của vùng cũ; low mới không tự tạo Spring.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E010 — SOS: 2026-10-21

- structure_id: HDG_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 15.6 vượt đỉnh hồi nhỏ 19/10=15.2 nhưng biên AR vùng mới 15.7 chưa hình thành.
- Bằng chứng volume: 2043600; RV20 2.75.
- Phản ứng đến cutoff: 22/10 high 15.7 rồi quay xuống close 15 ngày 30/10.
- Bằng chứng phản đối/giới hạn: Có thể là breakout vùng rất nhỏ nếu định nghĩa structure_id khác; chưa đủ xem là SOS của toàn vùng tháng 10.
- Điều kiện xác nhận: Cần TR có biên trước sự kiện và breakout giữ được; hiện ưu tiên E007 AR.
- Điều kiện bác bỏ: Nếu chỉ là nhịp AR không giữ breakout, không xác nhận SOS vùng tháng 10.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| Cụm đáy 27–28/07 RV20 dưới 1 | Có phản ứng hồi nhưng chưa phải climax volume | Không ép SC vì là đáy |
| 04/09 RV20 3.72 rồi giá trở lại vùng | Breakout thiếu duy trì | Không dùng volume làm tiêu chí duy nhất |
| 02–05/10 đóng dưới 15.2 | Tiếp tục markdown sau Accumulation thất bại | Không gán SOW của phân phối/tái phân phối |
| 21/10 RV20 2.75, close ở high | Có cầu trên nhịp hồi mới | Có thể chỉ là AR, chưa có SOS của TR mới |
| 30/10 close=low=15 | Pullback chưa xác nhận cầu giữ vùng | Không tự gọi ST hoặc LPS thành công |

Không gán một pha cho cả hai vùng. Giả thuyết tích lũy cũ đã suy yếu; vùng mới cần thêm kiểm tra. Nếu về sau có SOS mới, phải gắn đúng vùng và biên có trước sự kiện.

## 7. Kịch bản

- **S1: Thất bại vùng cũ, markdown tiếp diễn.** structure_id: HDG_TR_202607. Xác nhận: Không giữ lại 15.2–15.6 và phá 14.15 với follow-through; các nhịp hồi bị từ chối. Bác bỏ: Lấy lại bền 17.7 rồi kiểm tra giữ 17.2–17.7 phản đối cấu trúc suy yếu cũ. Bằng chứng nền: E004, E005.

- **S2: Xây nền tích lũy mới.** structure_id: HDG_TR_202610. Xác nhận: Giữ đáy 14.15, có kiểm tra cung giảm và phản ứng cầu; sau đó hai close trên 15.7 cùng follow-through và kiểm tra giữ vùng. Bác bỏ: Mất 14.15 qua hai close và không lấy lại trong ba phiên tiếp theo. Bằng chứng nền: E006, E007, E008.

- **S3: Dao động chưa rõ trong 14.15–15.7.** structure_id: HDG_TR_202610. Xác nhận: Tiếp tục giao dịch bên trong vùng, chưa có thoát vùng giữ được. Bác bỏ: Breakout hoặc breakdown bền vững. Bằng chứng nền: E006, E007, E008.

- **S4: Phục hồi vào vùng cũ.** structure_id: HDG_TR_202607. Xác nhận: Lấy lại 15.6 và giữ qua kiểm tra là cải thiện đầu tiên; vượt 17.7 bền mới là bằng chứng mạnh thoát vùng rộng. Bác bỏ: Bị từ chối tại 15.2–15.6 và tiếp tục tạo đáy thấp. Bằng chứng nền: E005, E007.

Ngưỡng hai/ba phiên là minh họa, chưa backtest. Không gán xác suất hoặc mục tiêu giá thiếu cơ sở.

## 8. Duyệt mẫu

- Trạng thái: in_review; eligible_for_ai=false.
- Người duyệt/ngày duyệt: chưa điền. Chưa có nhãn đúng/sai được người dùng xác nhận.
- Cần duyệt vùng cũ: SC uncertain, AR, ST và SOS 04/09 có phản ứng thất bại; E005 đã sửa thành tiếp tục markdown sau khi Accumulation thất bại.
- Cần duyệt vùng mới: stopping action/SC 09/10, AR đến 22/10, ST nông cuối tháng và SOS 21/10 còn uncertain.
- Đề xuất bác bỏ Spring thành công 09/10 của vùng cũ do không phục hồi vào vùng nhanh.
- Đối chiếu ảnh và điều chỉnh giá; sửa final_label, label_verdict, correction_reason và reviewer trong JSON cùng báo cáo.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Không có dữ liệu sau 30/10/2026. evaluation_date=null và observations=[]; các kịch bản chưa phải diễn biến đã xảy ra.
