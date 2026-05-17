export function formatDate(value) {
  return new Intl.DateTimeFormat("zh-CN", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

export function splitTags(value = "") {
  return value
    .split(",")
    .map((tag) => tag.trim())
    .filter(Boolean);
}
