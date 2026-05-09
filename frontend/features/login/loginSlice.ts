import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type UserRole = "client" | "freelancer" | "admin" | "unknown";

interface AuthUser {
  userId?: string;
  email?: string;
  firstName?: string;
  lastName?: string;
}

interface LoginState {
  isAuthenticated: boolean;
  isHydrated: boolean;
  accessToken: string | null;
  expiresAt: number | null;
  userRole: UserRole;
  user: AuthUser | null;
}

const initialState: LoginState = {
  isAuthenticated: false,
  isHydrated: false,
  accessToken: null,
  expiresAt: null,
  userRole: "unknown",
  user: null,
};

const loginSlice = createSlice({
  name: "login",
  initialState,
  reducers: {
    setAuthState(
      state,
      action: PayloadAction<{
        accessToken: string;
        expiresAt: number;
        userRole: UserRole;
        user?: AuthUser | null;
      }>
    ) {
      state.isAuthenticated = true;
      state.isHydrated = true;
      state.accessToken = action.payload.accessToken;
      state.expiresAt = action.payload.expiresAt;
      state.userRole = action.payload.userRole;
      state.user = action.payload.user ?? null;
    },
    updateAccessToken(
      state,
      action: PayloadAction<{ accessToken: string; expiresAt: number }>
    ) {
      state.isAuthenticated = true;
      state.isHydrated = true;
      state.accessToken = action.payload.accessToken;
      state.expiresAt = action.payload.expiresAt;
    },
    updateAuthUser(state, action: PayloadAction<AuthUser | null>) {
      state.user = action.payload;
    },
    clearAuthState(state) {
      state.isAuthenticated = false;
      state.isHydrated = true;
      state.accessToken = null;
      state.expiresAt = null;
      state.userRole = "unknown";
      state.user = null;
    },
    markAuthHydrated(state) {
      state.isHydrated = true;
    },
  },
});

export const {
  setAuthState,
  updateAccessToken,
  updateAuthUser,
  clearAuthState,
  markAuthHydrated,
} =
  loginSlice.actions;

export const selectIsLoggedIn = (state: { login: LoginState }) =>
  state.login.isAuthenticated;
export const selectIsAuthenticated = (state: { login: LoginState }) =>
  state.login.isAuthenticated;
export const selectIsHydrated = (state: { login: LoginState }) =>
  state.login.isHydrated;
export const selectAccessToken = (state: { login: LoginState }) =>
  state.login.accessToken;
export const selectAccessTokenExpiresAt = (state: { login: LoginState }) =>
  state.login.expiresAt;
export const selectUserRole = (state: { login: LoginState }) =>
  state.login.userRole;
export const selectAuthUser = (state: { login: LoginState }) => state.login.user;
export const selectIsClient = (state: { login: LoginState }) =>
  state.login.userRole === "client";
export const selectIsFreelancer = (state: { login: LoginState }) =>
  state.login.userRole === "freelancer";

export default loginSlice.reducer;
