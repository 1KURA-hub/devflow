"""评论相关接口封装。"""

from typing import Optional

from clients.base import ApiResponse, BaseClient


class CommentClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def create(self, post_id: int, content: str) -> ApiResponse:
        return self.http.post(
            f"/api/posts/{post_id}/comments",
            json={"content": content},
        )

    def list(self, post_id: int, limit: Optional[int] = None) -> ApiResponse:
        params = {"limit": limit} if limit else None
        return self.http.get(f"/api/posts/{post_id}/comments", params=params)
