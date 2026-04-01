## Context

`jam-service` và `playback-service` đều cần xác thực `trackId` trước khi mutate trạng thái queue/playback, nhưng hiện chưa có contract catalog thống nhất cho lookup và playability. Điều này tạo ra logic kiểm tra rời rạc, khó tái sử dụng và có rủi ro không nhất quán về lỗi khi track không tồn tại hoặc không khả dụng.

Thay đổi này là cross-service (`catalog-service`, `jam-service`, `playback-service`) và yêu cầu hợp đồng dữ liệu ổn định cho pre-check command path.

## Goals / Non-Goals

**Goals:**
- Cung cấp API lookup theo `trackId` với metadata tối thiểu phục vụ quyết định playable/unavailable.
- Chuẩn hóa deterministic error cho `track_not_found` và `track_unavailable`.
- Bắt buộc queue và playback command path thực hiện catalog pre-check trước khi mutate state.
- Bổ sung contract tests để đảm bảo schema ổn định giữa producer/consumer.

**Non-Goals:**
- Thiết kế hệ recommendation/ranking từ catalog.
- Thay đổi business policy ngoài phạm vi tồn tại và khả dụng của track.
- Thêm cơ chế cache phức tạp đa tầng ở phase này.

## Decisions

1. Expose catalog validation contract
- Thêm endpoint lookup theo `trackId` ở `catalog-service` trả về `trackId`, playability state, reason code và metadata tối thiểu.
- Vì pre-check cần deterministic outcome, response luôn phân biệt rõ found/unavailable/not_found.
- Alternative cân nhắc: dùng event-driven cache-only trong command services. Loại bỏ vì tăng độ trễ đồng bộ dữ liệu và khó đảm bảo nhất quán ngay lập tức.

2. Centralized validation adapter in command services
- `jam-service` và `playback-service` dùng adapter/client chung cho catalog lookup để tránh ad-hoc logic.
- Khi catalog trả `not_found` hoặc `unavailable`, command bị reject trước mutation.
- Alternative cân nhắc: mỗi service tự gọi catalog theo cách riêng. Loại bỏ vì làm lệch contract và error mapping.

3. Deterministic error mapping
- Map `not_found` -> `track_not_found`, `unavailable` -> `track_unavailable` ở API layer của command services.
- Đảm bảo rejected command không mutate queue/playback state và không publish event side effects.

4. Test strategy
- Thêm contract tests cho schema catalog lookup.
- Thêm integration tests cho queue/playback pre-check path để xác nhận reject-before-mutation và error mapping.

## Risks / Trade-offs

- [Risk] Catalog lookup làm tăng latency command path. -> Mitigation: timeout ngắn, fail-fast, và khả năng thêm cache ở phase sau.
- [Risk] Catalog/service contract drift gây lỗi tích hợp. -> Mitigation: contract tests bắt buộc trong CI cho 3 service.
- [Risk] Spike traffic gây tăng tải catalog-service. -> Mitigation: giới hạn timeout, backpressure, và rollout theo canary.

## Migration Plan

1. Triển khai endpoint lookup contract ở `catalog-service` và xác nhận schema.
2. Tích hợp pre-check vào `jam-service` queue command path sau feature toggle.
3. Tích hợp pre-check vào `playback-service` command path sau feature toggle.
4. Chạy contract/integration test, sau đó rollout canary và theo dõi tỷ lệ reject.
5. Mở rộng rollout toàn phần khi latency và error profile ổn định.

Rollback:
- Tắt feature toggle pre-check ở command services để quay về flow cũ.
- Giữ catalog endpoint để phục vụ verification và re-enable nhanh.

## Open Questions

- Metadata tối thiểu nào bắt buộc trong phase 1 ngoài trạng thái playable?
- Timeout tối ưu cho catalog lookup trong queue vs playback path là bao nhiêu?
- Cần chuẩn hóa error payload chi tiết (ví dụ reason subcode) đến mức nào ở phase này?
