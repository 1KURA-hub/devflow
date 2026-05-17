import React, { createContext, useContext, useEffect, useMemo, useState } from "react";
import { api, getStoredToken, setStoredToken } from "../api/client";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(Boolean(getStoredToken()));

  useEffect(() => {
    let active = true;
    if (!getStoredToken()) {
      setLoading(false);
      return undefined;
    }
    api
      .me()
      .then((nextUser) => {
        if (active) {
          setUser(nextUser);
        }
      })
      .catch(() => {
        if (active) {
          setStoredToken(null);
          setUser(null);
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, []);

  const value = useMemo(
    () => ({
      user,
      loading,
      async login(form) {
        const result = await api.login(form);
        setStoredToken(result.token);
        setUser(result.user);
        return result;
      },
      async register(form) {
        const result = await api.register(form);
        setStoredToken(result.token);
        setUser(result.user);
        return result;
      },
      logout() {
        setStoredToken(null);
        setUser(null);
      }
    }),
    [loading, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}
