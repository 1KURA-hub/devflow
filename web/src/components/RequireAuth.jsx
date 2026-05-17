import React from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../state/auth";

export function RequireAuth({ children }) {
  const { user, loading } = useAuth();
  if (loading) {
    return <div className="surface state-box">正在确认登录状态...</div>;
  }
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  return children;
}
