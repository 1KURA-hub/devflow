"""通知相关接口封装。"""

from clients.base import ApiResponse, BaseClient


class NotificationClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def list(self, limit=None, cursor=None) -> ApiResponse:
        params = {}
        if limit:
            params["limit"] = limit
        if cursor:
            params["cursor"] = cursor
        return self.http.get("/api/notifications", params=params or None)

    def unread_count(self) -> ApiResponse:
        return self.http.get("/api/notifications/unread-count")

    def mark_read(self, notification_id: int) -> ApiResponse:
        return self.http.post(f"/api/notifications/{notification_id}/read")

    def mark_all_read(self) -> ApiResponse:
        return self.http.post("/api/notifications/read-all")
