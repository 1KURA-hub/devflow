"""帖子相关接口封装。"""

from typing import Optional

from clients.base import ApiResponse, BaseClient


class PostClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def create(
        self,
        title: str,
        content: str,
        cover_url: str = "",
        tags: str = "",
    ) -> ApiResponse:
        return self.http.post(
            "/api/posts",
            json={
                "title": title,
                "content": content,
                "cover_url": cover_url,
                "tags": tags,
            },
        )

    def get(self, post_id: int) -> ApiResponse:
        return self.http.get(f"/api/posts/{post_id}")

    def delete(self, post_id: int) -> ApiResponse:
        return self.http.delete(f"/api/posts/{post_id}")

    def list_by_user(
        self,
        user_id: int,
        limit: Optional[int] = None,
        cursor: Optional[str] = None,
    ) -> ApiResponse:
        params = {}
        if limit:
            params["limit"] = limit
        if cursor:
            params["cursor"] = cursor
        return self.http.get(f"/api/users/{user_id}/posts", params=params)
