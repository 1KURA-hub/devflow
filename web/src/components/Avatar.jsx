import React from "react";
import guestAvatar from "../assets/guest-avatar.png";

export function Avatar({ user, label = "D", className = "profile-avatar" }) {
  const name = user?.nickname || label;
  const avatarURL = user?.avatar_url || "";
  if (avatarURL) {
    return <img className={`${className} avatar-image`} src={avatarURL} alt={name} />;
  }
  return <img className={`${className} avatar-image default-avatar-image`} src={guestAvatar} alt={name} />;
}
