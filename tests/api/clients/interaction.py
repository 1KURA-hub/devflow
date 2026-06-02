"""点赞 / 收藏 接口封装。"""

from clients.base import ApiResponse, BaseClient


class InteractionClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def like(self, post_id: int) -> ApiResponse:
        return self.http.post(f"/api/posts/{post_id}/like")

    def unlike(self, post_id: int) -> ApiResponse:
        return self.http.delete(f"/api/posts/{post_id}/like")

    def favorite(self, post_id: int) -> ApiResponse:
        return self.http.post(f"/api/posts/{post_id}/favorite")

    def unfavorite(self, post_id: int) -> ApiResponse:
        return self.http.delete(f"/api/posts/{post_id}/favorite")
