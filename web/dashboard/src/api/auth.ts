import { apiClient } from "./client";
import type { LoginRequest, SignupRequest, SignupResponse, LoginResponse } from "@/types";

export const authApi = {
  login: (data: LoginRequest) =>
    apiClient.post<LoginResponse>("/v1/auth/login", data),

  signup: (data: SignupRequest) =>
    apiClient.post<SignupResponse>("/v1/auth/signup", data),

  resendVerification: (email: string) =>
    apiClient.post<{ message: string }>("/v1/auth/resend-verification", { email }),

  logout: () => {
    apiClient.clearToken();
  },

  getCurrentUser: () => {
    // Decode JWT to get user info
    const token = apiClient.getToken();
    if (!token) return null;

    try {
      const base64Url = token.split(".")[1];
      const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split("")
          .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
          .join("")
      );
      return JSON.parse(jsonPayload);
    } catch {
      return null;
    }
  },
};
