import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./components/AppShell";
import { RequireAuth } from "./components/RequireAuth";
import { AuthPage } from "./pages/AuthPage";
import { FavoritesPage } from "./pages/FavoritesPage";
import { FeedPage } from "./pages/FeedPage";
import { MobileMePage } from "./pages/MobileMePage";
import { NotificationsPage } from "./pages/NotificationsPage";
import { PostDetailPage } from "./pages/PostDetailPage";
import { ProfilePage } from "./pages/ProfilePage";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<AuthPage mode="login" />} />
      <Route path="/register" element={<AuthPage mode="register" />} />
      <Route element={<AppShell />}>
        <Route path="/" element={<FeedPage mode="latest" />} />
        <Route path="/hot" element={<FeedPage mode="hot" />} />
        <Route
          path="/following"
          element={
            <RequireAuth>
              <FeedPage mode="following" />
            </RequireAuth>
          }
        />
        <Route path="/post/:id" element={<PostDetailPage />} />
        <Route path="/me" element={<MobileMePage />} />
        <Route path="/user/:id" element={<ProfilePage />} />
        <Route
          path="/favorites"
          element={
            <RequireAuth>
              <FavoritesPage />
            </RequireAuth>
          }
        />
        <Route
          path="/notifications"
          element={
            <RequireAuth>
              <NotificationsPage />
            </RequireAuth>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
