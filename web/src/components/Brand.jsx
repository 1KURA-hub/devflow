import React from "react";
import { Link } from "react-router-dom";

export function Brand() {
  return (
    <Link className="brand" to="/">
      <span className="brand-mark" aria-hidden="true">
        <svg viewBox="0 0 32 32">
          <path d="M8 10.5c2.3-3.1 5.2-4.7 8.7-4.7 3.1 0 5.7 1.2 7.8 3.6" />
          <path d="M7.5 16c2.4-2.6 5.1-3.9 8.2-3.9 3.9 0 6.9 1.6 8.8 4.8" />
          <path d="M8 21.8c2.2 2.8 5 4.2 8.4 4.2 3.1 0 5.7-1.2 7.8-3.6" />
        </svg>
      </span>
      <span className="brand-word">DevFlow</span>
    </Link>
  );
}
