"""三种 feed 接口封装。"""

from typing import Optional

from clients.base import ApiResponse, BaseClient


class FeedClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def latest(self, limit: Optional[int] = None) -> ApiResponse:
        params = {"limit": limit} if limit else None
        return self.http.get("/api/feed/latest", params=params)

    def hot(self, limit: Optional[int] = None) -> ApiResponse:
        params = {"limit": limit} if limit else None
        return self.http.get("/api/feed/hot", params=params)

    def following(self, limit: Optional[int] = None) -> ApiResponse:
        params = {"limit": limit} if limit else None
        return self.http.get("/api/feed/following", params=params)
