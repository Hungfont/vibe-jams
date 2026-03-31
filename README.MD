jam-service
Service lõi cho tính năng Jam: tạo/kết thúc session, join bằng token, quản lý queue, phân quyền host/guest, moderation, phát sự kiện realtime (jam.queue.updated, jam.playback.updated, jam.ended).
Liên quan mạnh tới Phase 1 và Phase 2.

playback-service
Xử lý điều khiển phát nhạc (play/pause/next/prev/seek) và đảm bảo đồng bộ trạng thái playback giữa các client (version/epoch, chống lệch trạng thái).
Trọng tâm ở US-P1-004, US-P1-005, US-P2-004.

auth-service
Xác thực token, cung cấp identity claims ổn định, kiểm tra entitlement Premium cho host, hỗ trợ cơ chế rate limit/abuse theo danh tính ở phase hardening.
Trọng tâm ở US-P1-007 và US-P4-004.

catalog-service
Tra cứu metadata bài hát, kiểm tra track có tồn tại/khả dụng, áp policy truy cập catalog cho queue/recommendation, và có cơ chế fallback khi sự cố để giữ hệ thống chạy.
Trọng tâm ở US-P1-008, US-P2-006, US-P3-005, US-P4-005.

api-service
Cổng API hướng client (entrypoint): nhận request từ app, validate quyền/scope cho endpoint recommendation theo session/user, phối hợp với service khác để trả response đầy đủ cho client.
Trọng tâm ở US-P3-004 và vai trò điều phối chung.