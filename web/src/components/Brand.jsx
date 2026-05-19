import React from "react";
import { Link } from "react-router-dom";
import devflowIcon from "../assets/devflow-icon.png";

export function Brand() {
  return (
    <Link className="brand" to="/">
      <span className="brand-mark" aria-hidden="true">
        <img src={devflowIcon} alt="" />
      </span>
      <span className="brand-word">DevFlow</span>
    </Link>
  );
}
