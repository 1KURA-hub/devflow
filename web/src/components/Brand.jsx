import React from "react";
import { Link } from "react-router-dom";

export function Brand() {
  return (
    <Link className="brand" to="/">
      <span className="brand-mark">D</span>
      <span className="brand-word">DevFlow</span>
    </Link>
  );
}
