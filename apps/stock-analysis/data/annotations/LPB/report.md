# LPB — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn đã được người dùng review và xác nhận ngày 31/10/2026. review.status=approved; eligible_for_ai=true. Phân tích từ OHLCV, chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_LPB_1D.csv](candles_LPB_1D.csv): 497 phiên, 31/10/2024–30/10/2026; cột event trống.
- [LPB_nhan.png](LPB_nhan.png): ảnh tham chiếu để người duyệt kiểm tra.
- source_commit: 11831cec05429a357b3f06016a8f7ff898afd096.
- Close cuối 36.1 theo đơn vị CSV. Chưa xác minh đơn vị, điều chỉnh giá hoặc đủ phiên theo lịch thị trường.
- RV20=volume/trung bình 20 phiên liền trước, loại phiên hiện tại.

## 2. Kết luận nhanh

**Ưu tiên markdown, pha E, sau vùng phân phối ứng viên 24/06–03/09/2026; tin cậy trung bình định tính.** Mất hỗ trợ 50.1–50.8 rồi hồi không lấy lại vùng là bằng chứng chính. Ngày 30/10 đóng ở đáy 36.1, phá 37.2; chưa có stopping action hay tích lũy mới được xác nhận.

## 3. Bối cảnh

- Close 28/04=43.09; high 25/06=58.5 sau nhịp tăng mạnh 19–25/06.\n- Sau đỉnh tháng 6, các lần hồi tháng 7 chỉ đạt high 56, sau đó phá vùng 50.1–50.8.\n- Close 30/10=36.1, dưới low 14–15/10=37.2; chuỗi markdown chưa có dấu hiệu dừng giảm được xác nhận.

Không suy đoán nguyên nhân doanh nghiệp hoặc ý định tổ chức từ giá/volume. Chưa thêm nhận định ngành/vĩ mô vì thiếu dữ liệu trong mẫu.

## 4. Cấu trúc và Trading Range

| structure_id | Thời gian | Hỗ trợ | Kháng cự | Trạng thái |
|---|---|---|---|---|
| LPB_TR_202606 | 24/06–03/09 | 50.1–50.8 | 56–58.5 | Phân phối ứng viên, provisional |
| LPB_MARKDOWN_202609 | Từ 03/09 | Chưa có hỗ trợ mới được xác nhận | Theo dõi các đỉnh hồi | Markdown developing |

Cạnh dưới từ low 03/07=50.8 và 14/07=50.1; cạnh trên từ đỉnh 25/06=58.5 và các lần kiểm tra high 56 tháng 7. Ngày kết thúc TR là mốc breakdown ứng viên, không phải xác nhận của người dùng. 37.2 đã mất; 36.1 là đáy mới, chưa tự gọi là hỗ trợ.

## 5. Nhật ký nhãn sự kiện

Người dùng đã duyệt E001–E005 ở mức ứng viên và xác nhận bác bỏ E006–E008. Duyệt mẫu không tự chuyển candidate thành confirmed. AR trong phân phối là **Automatic Reaction**.

### E001 — BC: 2026-06-24 → 2026-06-25

- structure_id: LPB_TR_202606; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 24/06 O/H/L/C=56.2/56.2/54.8/55.5; 25/06 high 58.5 nhưng close 56. 25/06 close ở 16.7% biên độ; không đồng nhất ngày volume lớn nhất và ngày đỉnh giá.
- Bằng chứng volume: 24/06 volume 11865500 RV20 8.17; 25/06 volume 3744100 RV20 1.89.
- Phản ứng đã quan sát đến cutoff: 26/06 close 53; nhịp giảm sau đó chạm low 50.1 ngày 14/07.
- Bằng chứng phản đối/giới hạn: Volume cực lớn có thể có nguyên nhân khác; chưa xác minh điều chỉnh hoặc giao dịch đặc biệt.
- Điều kiện xác nhận: Climax sau tăng, xuất hiện phản ứng giảm và vùng kiểm tra về sau.
- Điều kiện bác bỏ: Nếu cấu trúc sau đó tiếp tục tăng bền qua 58.5, xem lại BC của vùng này.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: BC; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E002 — AR: 2026-06-26 → 2026-07-14

- structure_id: LPB_TR_202606; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Giảm sau BC: 03/07 low 50.8; 14/07 low 50.1, close 51.5.
- Bằng chứng volume: 14/07 volume 2294600 RV20 0.81.
- Phản ứng đã quan sát đến cutoff: Sau 14/07 hồi đến high 56 ngày 21/07 và 27/07.
- Bằng chứng phản đối/giới hạn: AR ở phân phối là Automatic Reaction, không phải Automatic Rally; phạm vi kéo dài cần duyệt.
- Điều kiện xác nhận: Phản ứng giảm sau BC tạo cạnh dưới để theo dõi TR.
- Điều kiện bác bỏ: Nếu BC/phạm vi vùng bị sửa thì đánh giá lại AR và cạnh dưới.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: AR; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E003 — ST: 2026-07-21 → 2026-07-27

- structure_id: LPB_TR_202606; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 21/07 và 27/07 high 56, thấp hơn BC high 58.5; close lần lượt 54.6 và 54.0.
- Bằng chứng volume: 21/07 volume 6305000 RV20 1.93; 27/07 5562100 RV20 2.12.
- Phản ứng đã quan sát đến cutoff: Sau đỉnh kiểm tra, giá không vượt 56 và giảm về hỗ trợ tháng 8.
- Bằng chứng phản đối/giới hạn: Volume thấp hơn đỉnh 24/06 nhưng vẫn lớn so trung bình gần nhất; không mô tả là test cạn cung.
- Điều kiện xác nhận: Kiểm tra vùng cung với đỉnh thấp hơn và không duy trì được nhịp tăng.
- Điều kiện bác bỏ: Nếu vượt bền 58.5 và giữ được, bác bỏ diễn giải ST yếu của phân phối.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: ST; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E004 — SOW: 2026-08-20 → 2026-09-03

- structure_id: LPB_TR_202606; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 20/08 low 49.65; 21/08 low 48.5, close 50; 03/09 low 48, close 48.8 dưới hỗ trợ 50.1–50.8.
- Bằng chứng volume: 21/08 volume 3868700 RV20 1.26; 03/09 volume 3233600 RV20 1.07.
- Phản ứng đã quan sát đến cutoff: Các nhịp hồi sau đó không lấy lại vùng bền vững; giá tiếp tục giảm trong tháng 9–10.
- Bằng chứng phản đối/giới hạn: Đây là cụm breakdown diễn ra qua nhiều phiên, không chọn một phiên volume lớn nhất; 21/08 chưa đủ xác nhận một mình.
- Điều kiện xác nhận: Giá chấp nhận dưới hỗ trợ và có follow-through giảm; xem cả kết quả thay vì chỉ volume.
- Điều kiện bác bỏ: Nhanh lấy lại và giữ 50.8 sau kiểm tra sẽ phản đối SOW thành công.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: SOW; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E005 — LPSY: 2026-09-07

- structure_id: LPB_TR_202606; nguồn: AI_OHLCV_example.
- Bằng chứng giá: High 50, close 49.65; nhịp hồi không lấy lại hỗ trợ cũ 50.1–50.8 sau breakdown 03/09.
- Bằng chứng volume: Volume 1913700 RV20 0.63.
- Phản ứng đã quan sát đến cutoff: 09/09 close 48.15; 11/09 46.7; 14/09 46; tiếp tục giảm.
- Bằng chứng phản đối/giới hạn: Khoảng hồi ngắn; chỉ là LPSY ứng viên của vùng cũ, không gán mọi đỉnh hồi sau đó là LPSY.
- Điều kiện xác nhận: Hồi yếu sau SOW, bị từ chối dưới hỗ trợ cũ và có nhịp giảm tiếp.
- Điều kiện bác bỏ: Nếu lấy lại 50.8 và giữ qua kiểm tra, xem lại LPSY thành công.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: LPSY; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E006 — SC: 2026-10-15

- structure_id: LPB_MARKDOWN_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low=close=37.2; volume tăng nhưng sau nhịp hồi, 30/10 low=close=36.1 tạo đáy mới.
- Bằng chứng volume: 15/10 volume 2613600 RV20 1.31; không phải volume climax cực đoan.
- Phản ứng đã quan sát đến cutoff: 23/10 high 41.7, close 40.4; sau đó giá quay giảm và phá 37.2.
- Bằng chứng phản đối/giới hạn: Không bác bỏ khả năng có selling climax trong tương lai; chỉ bác bỏ SC đã dừng markdown tại 15/10 trong mẫu này.
- Điều kiện xác nhận: Cần stopping action, AR và kiểm tra giữ đáy để củng cố SC.
- Điều kiện bác bỏ: Đáy mới 36.1 tại cutoff phản đối giả thuyết 15/10 đã dừng markdown.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Kết quả duyệt: Bác bỏ nhãn. no_event chỉ áp dụng nhãn và phạm vi cấu trúc này.

### E007 — SOS: 2026-10-23

- structure_id: LPB_MARKDOWN_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: High 41.7, close 40.4; đây là hồi trong markdown, không có TR tích lũy đã xác lập bị phá lên.
- Bằng chứng volume: Volume 3460900 RV20 1.75; volume hồi lớn riêng lẻ không đủ tạo SOS.
- Phản ứng đã quan sát đến cutoff: 27/10 close 38.4; 30/10 close 36.1.
- Bằng chứng phản đối/giới hạn: Không biến mọi nhịp hồi mạnh thành SOS; thiếu cấu trúc tích lũy nền.
- Điều kiện xác nhận: Cần bối cảnh tích lũy, breakout biên trên có trước và follow-through.
- Điều kiện bác bỏ: Thiếu TR phù hợp và giá phá đáy sau hồi: đề xuất bác bỏ SOS này.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Kết quả duyệt: Bác bỏ nhãn. no_event chỉ áp dụng nhãn và phạm vi cấu trúc này.

### E008 — SC: 2026-10-30

- structure_id: LPB_MARKDOWN_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=37.25/37.3/36.1/36.1; đóng đáy, phá 37.2.
- Bằng chứng volume: Volume 2352300 RV20 1.15, không cực đoan.
- Phản ứng đã quan sát đến cutoff: Không có dữ liệu sau 30/10.
- Bằng chứng phản đối/giới hạn: Chưa có phản ứng cầu hoặc AR sau phiên; không chọn SC chỉ vì đây là đáy mới nhất.
- Điều kiện xác nhận: Chờ stopping action và phản ứng sau đó; chỉ xem đây là điểm theo dõi, chưa khẳng định SC.
- Điều kiện bác bỏ: Nếu giá tiếp tục tạo đáy thấp hơn không có chuyển biến cung–cầu, bỏ SC ứng viên.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Kết quả duyệt: Bác bỏ nhãn. no_event chỉ áp dụng nhãn và phạm vi cấu trúc này.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 24/06 RV20 8.17; 25/06 tạo high 58.5 rồi close 56 | BC cụm ứng viên sau tăng | Không đồng nhất đỉnh giá và đỉnh volume |
| Tháng 7 kiểm tra high 56 thấp hơn 58.5 | Cầu không đẩy được lên đỉnh cũ | Nhãn ST cần đánh giá cấu trúc, không chỉ đỉnh thấp |
| 03/09 close 48.8; 07/09 hồi high 50, RV20 0.63 | SOW/LPSY ứng viên và cung chiếm ưu thế | Breakdown không có volume cực đoan |
| 23/10 hồi tới high 41.7 rồi 30/10 phá đáy | Chưa có stopping action bền | Không kết luận mọi nhịp hồi đều không có cầu |

Pha E diễn giải markdown của TR cũ. Không gán pha A tích lũy mới chỉ vì giá đã giảm sâu. Nếu xuất hiện stopping action về sau, phải tạo structure_id mới và dùng dữ liệu thực tế.

## 7. Kịch bản

- **S1: Markdown tiếp diễn.** structure_id: LPB_MARKDOWN_202609. Xác nhận: Không lấy lại 37.2, tiếp tục đỉnh/đáy thấp hơn và nhịp hồi bị bán xuống. Bác bỏ: Ngừng tạo đáy mới, hồi và kiểm tra giữ đáy; lấy lại 41.7 sẽ phản đối chuỗi giảm ngắn hạn, chưa đủ bác bỏ TR cũ. Bằng chứng nền: E005, E008.

- **S2: Xuất hiện stopping action và khả năng vùng tích lũy mới.** structure_id: LPB_MARKDOWN_202609. Xác nhận: Có phản ứng cầu sau đáy mới, AR và kiểm tra giữ đáy. Vượt 41.7 rồi kiểm tra giữ được sẽ bổ sung bằng chứng thay đổi hành vi; không tự gọi tích lũy chỉ vì hồi một phiên. Bác bỏ: Tiếp tục tạo đáy thấp hơn mà không có phản ứng cầu bền. Bằng chứng nền: E006, E007, E008.

- **S3: Phục hồi trở lại vùng phân phối cũ.** structure_id: LPB_TR_202606. Xác nhận: Lấy lại 50.8 qua ít nhất hai close rồi kiểm tra giữ vùng; vượt 58.5 bền vững sẽ phản đối mạnh mô hình phân phối cũ. Bác bỏ: Nhịp hồi bị từ chối dưới 50.1–50.8 và quay giảm. Bằng chứng nền: E004, E005.

Các ngưỡng hai phiên là quy ước minh họa, chưa backtest. Không gán xác suất hoặc mục tiêu giá dưới 36.1 khi chưa có cơ sở trong mẫu.

## 8. Duyệt mẫu

- Trạng thái: approved; eligible_for_ai=true.
- Người duyệt: trungvdn; ngày duyệt: 2026-10-31.
- Assessment: markdown, pha E sau vùng phân phối ứng viên 24/06–03/09/2026.
- E001 BC, E002 AR, E003 ST, E004 SOW và E005 LPSY: giữ nhãn ứng viên theo phân tích đã duyệt.
- Xác nhận bác bỏ E006 SC 15/10, E007 SOS 23/10 và E008 SC 30/10 theo chỉnh sửa JSON của người dùng. Không gán SC tự động cho đáy mới nhất.
- no_event chỉ áp dụng nhãn/phạm vi bị bác bỏ; không có nghĩa phiên không mang thông tin giá/volume.
- Giữ nguyên giới hạn ảnh, điều chỉnh dữ liệu và ngưỡng chưa backtest. Kịch bản tương lai vẫn là giả thuyết.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Không có dữ liệu sau 30/10; evaluation_date=null, observations=[]. Không tự khẳng định S1/S2/S3 đã xảy ra.
