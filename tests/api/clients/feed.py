"""三种 feed 接口封装。"""

from typing import Optional

from clients.base import ApiResponse, BaseClient


class FeedClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def latest(self, limit: Optional[int] = None, cursor: Optional[str] = None) -> ApiResponse:
        params = {}
        if limit:
            params["limit"] = limit
        if cursor:
            params["cursor"] = cursor
        return self.http.get("/api/feed/latest", params=params or None)

    def hot(self, limit: Optional[int] = None, cursor: Optional[str] = None) -> ApiResponse:
        params = {}
        if limit:
            params["limit"] = limit
        if cursor:
            params["cursor"] = cursor
        return self.http.get("/api/feed/hot", params=params or None)

    def following(self, limit: Optional[int] = None, cursor: Optional[str] = None) -> ApiResponse:
        params = {}
        if limit:
            params["limit"] = limit
        if cursor:
            params["cursor"] = cursor
        return self.http.get("/api/feed/following", params=params or None)
