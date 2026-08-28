"""通用 HTTP 调用与响应解析。

devflow 所有接口统一返回 {"code": int, "message": str, "data": any}（见 internal/response/response.go）。
- 成功：HTTP 200，code=0
- 失败：HTTP 4xx/5xx，code=status，message=错误描述

ApiResponse 把这套结构包成一个对象，让用例可以同时断言 HTTP 状态码 + 业务字段。
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Optional

import requests

from config import BASE_URL, REQUEST_TIMEOUT


@dataclass
class ApiResponse:
    status_code: int
    code: int
    message: str
    data: Any

    @property
    def ok(self) -> bool:
        return self.status_code == 200 and self.code == 0


class BaseClient:
    """轻量 HTTP 封装。

    - 持有一个 requests.Session，自动复用连接
    - set_token() 注入 Bearer token
    - request() 统一解析返回体为 ApiResponse
    """

    def __init__(self, base_url: str = BASE_URL):
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self._token: Optional[str] = None

    def set_token(self, token: Optional[str]) -> None:
        self._token = token
        if token:
            self.session.headers["Authorization"] = f"Bearer {token}"
        else:
            self.session.headers.pop("Authorization", None)

    def close(self) -> None:
        """释放连接池资源；由 fixture 或临时 client 的创建方调用。"""
        self.session.close()

    def request(
        self,
        method: str,
        path: str,
        *,
        json: Any = None,
        params: Optional[dict] = None,
    ) -> ApiResponse:
        url = f"{self.base_url}{path}"
        resp = self.session.request(
            method=method,
            url=url,
            json=json,
            params=params,
            timeout=REQUEST_TIMEOUT,
        )
        body: dict = {}
        if resp.content:
            try:
                parsed = resp.json()
                body = parsed if isinstance(parsed, dict) else {}
            except ValueError:
                body = {}
        return ApiResponse(
            status_code=resp.status_code,
            code=body.get("code", -1),
            message=body.get("message", ""),
            data=body.get("data"),
        )

    def get(self, path: str, **kwargs) -> ApiResponse:
        return self.request("GET", path, **kwargs)

    def post(self, path: str, **kwargs) -> ApiResponse:
        return self.request("POST", path, **kwargs)

    def patch(self, path: str, **kwargs) -> ApiResponse:
        return self.request("PATCH", path, **kwargs)

    def delete(self, path: str, **kwargs) -> ApiResponse:
        return self.request("DELETE", path, **kwargs)
