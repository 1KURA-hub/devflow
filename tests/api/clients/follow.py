"""关注相关接口封装。"""

from clients.base import ApiResponse, BaseClient


class FollowClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def follow(self, user_id: int) -> ApiResponse:
        return self.http.post(f"/api/users/{user_id}/follow")

    def unfollow(self, user_id: int) -> ApiResponse:
        return self.http.delete(f"/api/users/{user_id}/follow")

    def state(self, user_id: int) -> ApiResponse:
        return self.http.get(f"/api/users/{user_id}/follow-state")

    def list_following(self, user_id: int) -> ApiResponse:
        return self.http.get(f"/api/users/{user_id}/following")

    def list_followers(self, user_id: int) -> ApiResponse:
        return self.http.get(f"/api/users/{user_id}/followers")
