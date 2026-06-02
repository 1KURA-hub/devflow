"""测试配置：所有可变项集中在这里，便于切换环境。"""

import os


BASE_URL = os.environ.get("DEVFLOW_BASE_URL", "http://localhost:8080")

REQUEST_TIMEOUT = float(os.environ.get("DEVFLOW_TIMEOUT", "10"))

DEFAULT_PASSWORD = "devflow123"

NOTIFICATION_POLL_TIMEOUT = float(os.environ.get("DEVFLOW_NOTIFY_TIMEOUT", "5"))
NOTIFICATION_POLL_INTERVAL = 0.3
