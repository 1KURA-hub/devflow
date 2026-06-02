"""通知相关接口封装。"""

from clients.base import ApiResponse, BaseClient


class NotificationClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def list(self) -> ApiResponse:
        return self.http.get("/api/notifications")

    def unread_count(self) -> ApiResponse:
        return self.http.get("/api/notifications/unread-count")

    def mark_read(self, notification_id: int) -> ApiResponse:
        return self.http.post(f"/api/notifications/{notification_id}/read")

    def mark_all_read(self) -> ApiResponse:
        return self.http.post("/api/notifications/read-all")
