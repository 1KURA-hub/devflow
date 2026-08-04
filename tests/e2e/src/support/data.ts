import { randomUUID } from "node:crypto";
import { env } from "../config/env";

/**
 * 后端用户名最多 12 个字符，所以不能直接拼完整时间戳或 UUID。
 * u_ + 5 位时间片 + 5 位随机串 = 12 个字符。
 */
export function uniqueUsername(): string {
  const timePart = Date.now().toString(36).slice(-5);
  const randomPart = randomUUID().replaceAll("-", "").slice(0, 5);
  return `u_${timePart}${randomPart}`;
}

export function uniquePostTitle(prefix = "pw"): string {
  return `${prefix}_${Date.now().toString(36)}_${randomUUID().slice(0, 6)}`;
}

export function newUserInput() {
  const username = uniqueUsername();
  return {
    username,
    password: env.defaultPassword,
    nickname: `n_${username.slice(-8)}`
  };
}
