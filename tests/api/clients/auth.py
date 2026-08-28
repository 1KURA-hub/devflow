"""认证相关接口封装。"""

from clients.base import ApiResponse, BaseClient


class AuthClient:
    def __init__(self, http: BaseClient):
        self.http = http

    def register(self, username: str, password: str, nickname: str) -> ApiResponse:
        return self.http.post(
            "/api/auth/register",
            json={"username": username, "password": password, "nickname": nickname},
        )

    def login(self, username: str, password: str) -> ApiResponse:
        return self.http.post(
            "/api/auth/login",
            json={"username": username, "password": password},
        )

    def me(self) -> ApiResponse:
        return self.http.get("/api/me")

    def user_profile(self, user_id: int) -> ApiResponse:
        return self.http.get(f"/api/users/{user_id}")

    def update_me(self, **fields) -> ApiResponse:
        return self.http.patch("/api/me", json=fields)
