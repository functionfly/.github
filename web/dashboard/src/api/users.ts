import { apiClient } from "./client";
import type { PublicUserProfile } from "@/types";

export interface UpdateProfileRequest {
  name?: string;
  username?: string;
  companyName?: string;
  bio?: string;
}

export interface UpdateProfileResponse {
  message: string;
  user: {
    id: string;
    name: string;
    username?: string;
    companyName?: string;
    email: string;
    avatar?: string;
    updatedAt: string;
  };
}

export interface PasswordChangeRequest {
  currentPassword: string;
  newPassword: string;
}

export const usersApi = {
  /**
   * Get the public profile for a user by username.
   * Returns only safe public fields — no email, tenantId, or role.
   */
  getPublicProfile: (username: string) =>
    apiClient.get<PublicUserProfile>(`/v1/users/${encodeURIComponent(username)}`),

  /**
   * Get the current authenticated user's full profile.
   */
  getMe: () =>
    apiClient.get<UpdateProfileResponse["user"]>("/v1/users/me"),

  /**
   * Update the current authenticated user's profile.
   */
  updateMe: (data: UpdateProfileRequest) =>
    apiClient.patch<UpdateProfileResponse>("/v1/users/me", data),

  /**
   * Change the current user's password.
   */
  changePassword: (data: PasswordChangeRequest) =>
    apiClient.post<{ message: string }>("/v1/users/me/change-password", data),

  /**
   * Request a password reset email.
   */
  requestPasswordReset: (email: string) =>
    apiClient.post<{ message: string }>("/v1/auth/password-reset", { email }),

  /**
   * Confirm a password reset with token and new password.
   */
  confirmPasswordReset: (token: string, newPassword: string) =>
    apiClient.post<{ message: string }>("/v1/auth/password-reset/confirm", {
      token,
      newPassword,
    }),
};
