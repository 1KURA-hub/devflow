import React from "react";

export function Avatar({ user, label = "D", className = "profile-avatar" }) {
  const name = user?.nickname || label;
  const avatarURL = user?.avatar_url || "";
  if (avatarURL) {
    return <img className={`${className} avatar-image`} src={avatarURL} alt={name} />;
  }
  return <div className={className}>{name.slice(0, 1)}</div>;
}
