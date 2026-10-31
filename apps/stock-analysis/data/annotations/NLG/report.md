# NLG — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn đã được người dùng review và xác nhận ngày 31/10/2026. review.status=approved; eligible_for_ai=true. Phân tích từ OHLCV; chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_NLG_1D.csv](candles_NLG_1D.csv): 497 phiên từ 31/10/2024 đến 30/10/2026; cột event trống.
- [NLG_nhan.png](NLG_nhan.png): ảnh tham chiếu để người duyệt đối chiếu.
- source_commit: a86721517b4490cb9731326dcce97d3dee27f344.
- Close cuối 22 theo đơn vị CSV; đơn vị và điều chỉnh giá chưa xác minh.
- Không trùng ngày; OHLC hợp lệ, volume không âm. Chưa đối chiếu đủ phiên theo lịch thị trường.
- RV20 = volume / trung bình 20 phiên liền trước, không tính phiên hiện tại. CLV=(close-low)/(high-low), bỏ khi high=low.

## 2. Kết luận nhanh

**Cấu trúc chưa rõ; giả thuyết ưu tiên là tích lũy nhỏ thử chuyển C→D, tin cậy thấp.** Cụm 14–15/10 là Spring ứng viên của vùng nhỏ 21.3–22.5; 21/10 là SOS ứng viên. Tuy nhiên, 30/10 đóng ở đáy phiên 22, dưới vùng breakout: **chưa có LPS xác nhận**. Pha D không phải kết luận chắc chắn, và không áp dụng cho toàn bộ vùng tháng 7–9.

## 3. Bối cảnh và phạm vi

- Vùng bối cảnh rộng: low 28/07=19.15; hồi đến high 07/09=25.95; sau đó giảm đến low 28/09=21.35.\n- Vùng chính trong ví dụ là vùng nhỏ cuối tháng 9–10, không phải toàn bộ vùng tháng 7–9.\n- Cụm 14–15/10 xuyên 21.3 nhưng không xuyên 19.15; nhãn Spring chỉ có thể áp dụng cho vùng nhỏ đã chọn.

Không mặc định đây là tái tích lũy: cần bối cảnh xu hướng tăng lớn hơn để phân biệt tích lũy, tái tích lũy và hồi kỹ thuật. Chưa bổ sung nhận định ngành/vĩ mô vì không có dữ liệu tương ứng trong mẫu.

## 4. Vùng giao dịch

| Phạm vi | Hỗ trợ | Kháng cự | Trạng thái |
|---|---|---|---|
| NLG_TR_202609_SHORT, từ 28/09 | 21.3–21.4 | 22.2–22.5 | TR ngắn, provisional |
| NLG_CONTEXT_202607, từ 27/07 | 19.15–20.0 | 25.7–25.95 | TR lớn provisional; AR rộng còn uncertain |

Biên nhỏ có trước Spring: 28–29/09 low 21.35, 01/10 low 21.3; 08/10 high 22.2, 09/10 high 22.5. Sau cụm xuyên đáy, 21.0 là đáy Spring ứng viên; không hồi tố sửa biên trước sự kiện xuống 21.0. TR ngắn nên không ép có đủ PS/SC/AR/ST.

## 5. Nhật ký nhãn sự kiện

Người dùng đã duyệt: các nhãn candidate được giữ ở mức ứng viên; các nhãn uncertain chưa chốt; E007–E008 được xác nhận là nhãn sai trong phạm vi ghi rõ. Duyệt mẫu không tự chuyển trạng thái thị trường thành confirmed. Nhận định sử dụng kiến thức đến cutoff, không giả vờ biết phản ứng tương lai ngay tại ngày sự kiện.

### E001 — SC: 2026-09-28

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=22.45/22.45/21.35/21.65; close ở 27.3% biên độ sau nhịp giảm từ 25.95.
- Bằng chứng volume: Volume 2265200; RV20 1.24, không phải volume cực đoan.
- Phản ứng đã quan sát: 29/09 giữ low 21.35; 01/10 low 21.3; tiếp theo dao động quanh 21.3–22.5.
- Bằng chứng phản đối/giới hạn: Climax không rõ; nên cho phép no_event/stopping action thay vì ép SC.
- Điều kiện xác nhận: Cần xác định nhịp bán cuối và thay đổi hành vi, không chỉ chọn đáy.
- Điều kiện bác bỏ: Nếu không đủ bằng chứng climax thì bỏ SC nhưng vẫn giữ vùng đi ngang ứng viên.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E002 — AR: 2026-10-01 → 2026-10-09

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 01/10 close=high=21.85 sau low 21.3; 09/10 high 22.5, close 21.9.
- Bằng chứng volume: 01/10 volume 2687800 RV20 1.42; 09/10 volume 1248000 RV20 0.73.
- Phản ứng đã quan sát: Sau đỉnh 09/10, giá quay xuống cụm 14–15/10.
- Bằng chứng phản đối/giới hạn: SC còn uncertain; nếu SC không được duyệt, đây chỉ là nhịp hồi tạo biên trên, nhãn AR cần xét lại.
- Điều kiện xác nhận: Có hồi sau stopping action và phản ứng tạo cạnh trên.
- Điều kiện bác bỏ: Nếu vùng hoặc SC gốc bị bác bỏ, đánh giá lại nhãn AR.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: AR; người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E003 — Spring: 2026-10-14 → 2026-10-15

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 14/10 low 21.15, close 21.2 dưới hỗ trợ 21.3; 15/10 low 21.0 rồi close 21.4 trở lại vùng.
- Bằng chứng volume: 14/10 volume 1102700; 15/10 921400 RV20 0.64.
- Phản ứng đã quan sát: 16–20/10 low không thấp hơn 21.2; 21/10 vượt vùng phía trên.
- Bằng chứng phản đối/giới hạn: Chỉ là Spring của TR ngắn; 15/10 low vẫn cao hơn đáy bối cảnh rộng 19.15.
- Điều kiện xác nhận: Xuyên biên đã tồn tại trước sự kiện, hồi vào vùng rồi giữ đáy 21.0.
- Điều kiện bác bỏ: Nếu chấp nhận giá dưới 21.0 kéo dài, bác bỏ giả thuyết Spring thành công.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: Spring; người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E004 — Test: 2026-10-20

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 21.2 cao hơn Spring low 21.0; close 21.4; biên độ 0.3.
- Bằng chứng volume: Volume 1091400 RV20 0.79 nhưng cao hơn volume 15/10=921400 khoảng 18.5%.
- Phản ứng đã quan sát: 21/10 tăng mạnh close 22.85.
- Bằng chứng phản đối/giới hạn: Volume không giảm so phiên Spring; hồi sau đó không tự chứng minh đây là Test chất lượng.
- Điều kiện xác nhận: Đánh giá đồng thời biên độ, giữ đáy và nguồn cung; không chỉ so với RV20.
- Điều kiện bác bỏ: Nếu yêu cầu Test volume thấp hơn Spring thì nhãn này không đạt; có thể để no_event.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E005 — SOS: 2026-10-21

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=21.65/22.85/21.65/22.85; vượt 22.5 khoảng 1.56%, đóng ở đỉnh phiên.
- Bằng chứng volume: Volume 2765500; RV20 1.97. Ngày 22/10 volume 3140500 RV20 2.17.
- Phản ứng đã quan sát: 22–23/10 close 23.15, 23.3; nhưng 30/10 đóng 22 trở lại dưới 22.5.
- Bằng chứng phản đối/giới hạn: Chỉ là breakout TR nhỏ; chưa vượt đỉnh bối cảnh rộng 25.95.
- Điều kiện xác nhận: Cần giữ hoặc lấy lại 22.5, có follow-through và pullback được cầu hỗ trợ.
- Điều kiện bác bỏ: Hai close dưới 22.2 và không lấy lại 22.5 trong ba phiên sau làm suy yếu breakout; đến cutoff chưa đủ diễn biến sau 30/10.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: SOS; người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E006 — LPS: 2026-10-27 → 2026-10-30

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 28/10 low 22.45, close 23; 30/10 low=close=22, mất vùng breakout 22.2–22.5.
- Bằng chứng volume: 28/10 volume 904800 RV20 0.60; 30/10 1044600 RV20 0.75, tăng so 28/10.
- Phản ứng đã quan sát: Không có dữ liệu hồi sau 30/10.
- Bằng chứng phản đối/giới hạn: Close đáy phiên 30/10 và mất vùng breakout không phù hợp LPS đã xác nhận.
- Điều kiện xác nhận: Chờ hồi lấy lại 22.5 rồi kiểm tra giữ vùng với cung giảm.
- Điều kiện bác bỏ: Nếu tiếp tục giảm dưới vùng và không phục hồi, không xác nhận LPS.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E007 — Test: 2026-10-12

- structure_id: NLG_TR_202609_SHORT; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 21.75, close 21.95; ngày này xảy ra trước Spring 14–15/10.
- Bằng chứng volume: Volume 322900 RV20 0.20: volume thấp không sửa được sai thứ tự.
- Phản ứng đã quan sát: Sau đó 14–15/10 mới xuất hiện xuyên hỗ trợ.
- Bằng chứng phản đối/giới hạn: Chỉ bác bỏ nhãn Test sau Spring đối với cấu trúc này, không bác bỏ mọi loại kiểm tra cung.
- Điều kiện xác nhận: Test sau Spring phải xảy ra sau Spring và có quan hệ với sự kiện đó.
- Điều kiện bác bỏ: 12/10 trước 14–15/10 nên không phải Test của Spring này.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.\n- Ghi chú duyệt: Người dùng xác nhận bác bỏ Test sau Spring ngày 12/10 vì sai thứ tự; no_event chỉ đối với nhãn Test này.

### E008 — SOS: 2026-10-21

- structure_id: NLG_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 22.85 vẫn dưới đỉnh bối cảnh rộng 25.95.
- Bằng chứng volume: Volume lớn không thay thế điều kiện breakout khỏi đúng vùng.
- Phản ứng đã quan sát: Nhịp hồi cao nhất 26/10 high 24.5 vẫn dưới 25.95.
- Bằng chứng phản đối/giới hạn: Đây là ví dụ nhãn SOS sai ở phạm vi rộng; E005 vẫn là SOS ứng viên vùng nhỏ.
- Điều kiện xác nhận: Phải nêu đúng structure_id và biên cần vượt.
- Điều kiện bác bỏ: Không gán breakout trên 25.95 cho phiên đóng 22.85.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.\n- Ghi chú duyệt: Người dùng xác nhận bác bỏ SOS vùng lớn ngày 21/10; no_event chỉ trong NLG_CONTEXT_202607, E005 vẫn là SOS ứng viên vùng nhỏ.

### E009 — SC: 2026-07-27 → 2026-07-28

- structure_id: NLG_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 27/07 O/H/L/C=21.2/21.7/20/20; 28/07=19.8/20.8/19.15/20.6. 28/07 tạo đáy mới nhưng đóng ở 87.9% biên độ; chọn cụm hai phiên thay vì coi 27/07 là đáy cuối.
- Bằng chứng volume: 27/07 volume 2779500 RV20 1.71; 28/07 volume 2768600 RV20 1.61.
- Phản ứng đã quan sát: 29/07–07/09 hồi lên high 25.95; đến cutoff chưa xuyên lại 19.15.
- Bằng chứng phản đối/giới hạn: Volume tăng vừa phải, không cực đoan; cần kiểm tra stopping action và nhịp AR chứ không gán SC chỉ vì là đáy.
- Điều kiện xác nhận: Có nhịp hồi rõ và kiểm tra về sau giữ đáy 19.15.
- Điều kiện bác bỏ: Nếu cấu trúc không giữ đáy và markdown tiếp tục bền dưới 19.15, đánh giá lại SC của vùng rộng.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: SC; người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E010 — AR: 2026-07-29 → 2026-09-07

- structure_id: NLG_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Nhịp hồi từ sau SC tới high 25.95 ngày 07/09; 04/09 close 25.7. Cạnh trên rộng tham chiếu 25.7–25.95.
- Bằng chứng volume: 04/09 volume 3330100 RV20 1.59; 07/09 2319000 RV20 1.08.
- Phản ứng đã quan sát: Sau 07/09 giảm, về low 21.35 ngày 28/09 và low 21.0 ngày 15/10.
- Bằng chứng phản đối/giới hạn: Nhịp kéo dài và có các đỉnh trung gian 03/08 high 22.45, 05/08 high 22.55; AR sơ cấp có thể kết thúc sớm hơn. Không ép toàn bộ nhịp hồi là Automatic Rally đã xác nhận.
- Điều kiện xác nhận: Người duyệt xác nhận nhịp rộng này thuộc AR của cấu trúc lớn, hoặc tách AR sơ cấp và nhịp hồi mở rộng.
- Điều kiện bác bỏ: Nếu AR sơ cấp kết thúc đầu tháng 8, sửa phạm vi E010; khi đó phải đánh giá lại biên trên thay vì tự coi 25.95 là đỉnh AR.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

### E011 — ST: 2026-09-28 → 2026-10-15

- structure_id: NLG_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Kiểm tra sau đỉnh hồi: low 28/09=21.35, 01/10=21.3, 15/10=21.0. Đáy 21.0 cao hơn SC low 19.15 khoảng 9.66%; đây là ST nông/kiểm tra bên trong ứng viên.
- Bằng chứng volume: 28/09 volume 2265200 bằng 82.0% volume SC 28/07; 15/10 volume 921400 bằng 33.3% volume SC; RV20 15/10=0.64.
- Phản ứng đã quan sát: 21–23/10 hồi với close 22.85, 23.15, 23.3; 30/10 close 22 vẫn trên đáy cụm kiểm tra 21.0.
- Bằng chứng phản đối/giới hạn: Chưa kiểm tra sát 19.15; volume 01/10=2687800 gần volume SC 28/07. Không kết luận toàn bộ nhịp đã cạn cung.
- Điều kiện xác nhận: Đáy cao hơn SC, cung giảm tương đối và có phản ứng hồi; cần người duyệt xác nhận đây là ST của vùng rộng thay vì chỉ là pullback.
- Điều kiện bác bỏ: Nếu không có quan hệ kiểm tra với SC hoặc mất đáy 19.15 bền vững, bác bỏ ST thành công; mất 21.0 trước hết là suy yếu nhịp kiểm tra nông.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: ST; người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 15/10 low 21, close 21.4; RV20 0.64 | Xuyên biên nhỏ rồi hồi, Spring ứng viên | Chưa xuyên đáy rộng 19.15; TR ngắn |
| 20/10 volume 1,091,400 cao hơn 15/10 khoảng 18.5% | Test còn nghi ngờ | RV20 thấp không đồng nghĩa volume thấp hơn Spring |
| 21/10 close=high=22.85; RV20 1.97 | Nỗ lực tăng có kết quả vượt vùng nhỏ | Không phải breakout trên 25.95 |
| 26/10 high 24.5, close 23, CLV 6.25% | Cầu không giữ được thành quả trong phiên | Không suy ra danh tính hoặc ý định bên bán |
| 30/10 close=low=22; RV20 0.75 | Giá giảm dù volume thấp, cần theo dõi thiếu cầu/cung còn lại | Không gọi LPS đã xác nhận chỉ vì volume giảm |

Pha D chỉ là giả thuyết đang bị kiểm tra. Nếu mất breakout và không lấy lại, hạ về dao động chưa rõ pha hoặc hồi kỹ thuật; không duy trì nhãn D bất chấp bằng chứng mới.

Với **NLG_CONTEXT_202607**, SC E009 → AR E010 → ST E011 là chuỗi giả thuyết riêng. Chưa có breakout trên 25.95 để gọi SOS vùng lớn; chưa gán pha D cho vùng lớn. Cụm 14–15/10 có thể đồng thời là phần ST của vùng lớn và Spring ứng viên của vùng nhỏ, không mâu thuẫn nếu giữ đúng structure_id. E010 cần duyệt đặc biệt: nhịp hồi dài có thể gồm AR sơ cấp đầu tháng 8 và các nhịp hồi mở rộng sau đó.

## 7. Kịch bản

### Vùng nhỏ — NLG_TR_202609_SHORT

- **S1: Giữ được breakout vùng nhỏ, củng cố C→D.** Xác nhận: Lấy lại 22.5 và giữ qua ít nhất hai close, pullback sau đó giữ vùng với cung giảm. Bác bỏ: Giảm tiếp, không lấy lại vùng; mất 21.0 làm suy yếu Spring.

- **S2: Breakout thất bại, quay lại dao động.** Xác nhận: Giá ở dưới 22.2 nhưng vẫn giữ 21.0; chưa có đà theo một hướng. Bác bỏ: Thoát vùng bền vững.

- **S3: Markdown tiếp diễn.** Xác nhận: Hai close liên tiếp dưới 21.0 và không hồi trong ba phiên sau. Bác bỏ: Hồi nhanh lấy lại vùng và kiểm tra giữ đáy.

### Vùng lớn — NLG_CONTEXT_202607

Mốc chính: hỗ trợ 19.15–20.0; đáy kiểm tra nông 21.0; kháng cự 25.7–25.95. **Mất 21.0 chưa đồng nghĩa mất vùng lớn; vượt 22.5 chưa phải SOS vùng lớn.** Cạnh trên còn phụ thuộc kết quả duyệt AR E010; nếu AR được sửa phải đánh giá lại các kịch bản này.

- **S4: Củng cố tích lũy vùng lớn và breakout phía trên.**
  - structure_id: NLG_CONTEXT_202607; status: hypothesis.
  - Điều kiện xác nhận: Giữ được đáy kiểm tra nông 21.0, phục hồi qua vùng nhỏ 22.5 rồi tiến đến 25.7–25.95. Chỉ xem xét SOS vùng lớn khi có ít nhất hai close trên 25.95, lực tăng/volume hỗ trợ và kiểm tra lại vùng breakout giữ được.
  - Điều kiện bác bỏ: Breakout trên 25.95 nhanh chóng quay xuống và không lấy lại vùng; hoặc mất bền đáy SC 19.15. Mất 21.0 chỉ làm suy yếu đường đi ST nông, chưa tự bác bỏ toàn bộ tích lũy lớn.
  - Bằng chứng nền: E009, E010, E011; mốc theo dõi: 21, 22.5, 25.7, 25.95.

- **S5: Tiếp tục xây nền và kiểm tra sâu trong vùng lớn.**
  - structure_id: NLG_CONTEXT_202607; status: hypothesis.
  - Điều kiện xác nhận: Không vượt bền 25.95; giá tiếp tục dao động trong 19.15–25.95. Nếu mất 21.0, theo dõi phản ứng tại 19.15–20.0: giữ hoặc phục hồi vùng này, volume bán giảm tương đối và xuất hiện nhịp hồi sẽ hỗ trợ giả thuyết kiểm tra sâu.
  - Điều kiện bác bỏ: Thoát bền phía trên 25.95 hoặc có hai close dưới 19.15 và không lấy lại trong ba phiên sau; khi đó không còn là kịch bản dao động giữ vùng lớn.
  - Bằng chứng nền: E009, E010, E011; mốc theo dõi: 19.15, 20, 21, 25.95.

- **S6: Xuyên đáy vùng lớn rồi phục hồi — Spring mới chỉ là khả năng.**
  - structure_id: NLG_CONTEXT_202607; status: hypothesis.
  - Điều kiện xác nhận: Sau khi vùng lớn được người duyệt chấp nhận, giá xuyên 19.15 rồi close trở lại trên 19.15 trong tối đa ba phiên; tiếp theo kiểm tra giữ đáy xuyên thủng với cung giảm và có phản ứng cầu. Phải có dữ liệu thực tế mới tạo nhãn Spring/Test vùng lớn.
  - Điều kiện bác bỏ: Giá ở dưới 19.15 kéo dài, hoặc kiểm tra sau hồi phá tiếp đáy xuyên thủng và không phục hồi. Không gán Spring chỉ vì một phiên xuyên đáy.
  - Bằng chứng nền: E009, E011; mốc theo dõi: 19.15, 20.

- **S7: Vùng lớn thất bại, markdown tiếp diễn.**
  - structure_id: NLG_CONTEXT_202607; status: hypothesis.
  - Điều kiện xác nhận: Mất 21.0 là cảnh báo sớm. Hai close liên tiếp dưới 19.15 và không lấy lại trong ba phiên sau xác nhận thất bại theo quy ước vận hành của mẫu; nhịp hồi bị từ chối dưới 19.15–20.0, đỉnh/đáy thấp hơn và cung tăng sẽ củng cố.
  - Điều kiện bác bỏ: Lấy lại 19.15–20.0 nhanh chóng, kiểm tra giữ được và có cầu hỗ trợ; chuyển sang đánh giá S6 thay vì tiếp tục khẳng định breakdown.
  - Bằng chứng nền: E009, E010, E011; mốc theo dõi: 19.15, 20, 21.

Các kịch bản vùng lớn mô tả các đường đi thay thế có thể chuyển tiếp, không phải xác suất hay diễn biến đã xảy ra. S6 và S7 được phân biệt bằng phản ứng sau xuyên đáy và mốc thời gian quan sát; tại thời điểm chưa đủ phiên, giữ trạng thái chờ xác nhận.

Các ngưỡng hai/ba phiên là quy ước minh họa để điều kiện kiểm tra được, chưa backtest và không phải quy tắc Wyckoff bắt buộc.

## 8. Duyệt mẫu

- Trạng thái: approved; eligible_for_ai=true.
- Người duyệt: trungvdn@gmail.com; ngày duyệt: 2026-10-31.
- Phạm vi duyệt: NLG_TR_202609_SHORT, NLG_CONTEXT_202607 và các kịch bản S1–S7.
- Giữ nhãn ứng viên: E002 AR, E003 Spring, E005 SOS (vùng nhỏ); E009 SC và E011 ST (vùng lớn).
- Giữ chưa chắc chắn: E001 SC vùng nhỏ, E004 Test, E006 LPS và E010 AR rộng. AR rộng chưa được tự thu hẹp hoặc chuyển confirmed.
- Xác nhận bác bỏ E007 Test 12/10 sai thứ tự và E008 SOS 21/10 sai phạm vi vùng lớn. no_event chỉ áp dụng nhãn/phạm vi bị bác bỏ.
- Duyệt theo xác nhận người dùng; giữ nguyên giới hạn xác minh ảnh, điều chỉnh giá và ngưỡng chưa backtest. Các kịch bản tương lai vẫn là giả thuyết.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Chưa có dữ liệu sau 30/10/2026. Giữ evaluation_date=null, observations=[]; không tự kết luận các kịch bản đã xảy ra.
