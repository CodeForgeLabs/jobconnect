"use client";

import { USER_ROUTES } from "@/lib/apiConfig";
import { apiClient } from "@/lib/apiClient";

export const userApi = {
  getProfile: () => apiClient.get(USER_ROUTES.profile, { auth: true }),
  patchProfile: (payload: Record<string, unknown>) =>
    apiClient.patch(USER_ROUTES.profile, payload, { auth: true }),
  deleteProfile: (hardDelete: boolean) =>
    apiClient.delete(`${USER_ROUTES.profile}?hard_delete=${hardDelete}`, { auth: true }),

  getOnboardingStatus: () => apiClient.get(USER_ROUTES.onboardingStatus, { auth: true }),

  getSettings: () => apiClient.get(USER_ROUTES.settings, { auth: true }),
  patchSettings: (payload: Record<string, unknown>) =>
    apiClient.patch(USER_ROUTES.settings, payload, { auth: true }),

  requestAvatarUploadUrl: (fileName: string, contentType: string) =>
    apiClient.post<{ storage_key: string; upload_url: string }>(
      USER_ROUTES.avatarUploadUrl,
      { file_name: fileName, content_type: contentType },
      { auth: true }
    ),
  upsertAvatar: (payload: Record<string, unknown>) =>
    apiClient.post(USER_ROUTES.avatar, payload, { auth: true }),
  getAvatar: () => apiClient.get(USER_ROUTES.avatar, { auth: true }),
  removeAvatar: () => apiClient.delete(USER_ROUTES.avatar, { auth: true }),

  requestCVUploadUrl: (fileName: string, contentType: string) =>
    apiClient.post<{ storage_key: string; upload_url: string }>(
      USER_ROUTES.cvUploadUrl,
      { file_name: fileName, content_type: contentType },
      { auth: true }
    ),
  upsertCV: (payload: Record<string, unknown>) =>
    apiClient.post(USER_ROUTES.cv, payload, { auth: true }),
  getCV: () => apiClient.get(USER_ROUTES.cv, { auth: true }),
  removeCV: () => apiClient.delete(USER_ROUTES.cv, { auth: true }),

  requestPortfolioUploadUrl: (fileName: string, contentType: string) =>
    apiClient.post<{ storage_key: string; upload_url: string }>(
      USER_ROUTES.portfolioUploadUrl,
      { file_name: fileName, content_type: contentType },
      { auth: true }
    ),
  listPortfolio: (pageSize = 20, pageToken = "") =>
    apiClient.get(
      `${USER_ROUTES.portfolio}?page_size=${pageSize}&page_token=${encodeURIComponent(pageToken)}`,
      { auth: true }
    ),
  getPortfolioItem: (itemId: string) =>
    apiClient.get(`${USER_ROUTES.portfolio}/${itemId}`, { auth: true }),
  createPortfolioItem: (payload: Record<string, unknown>) =>
    apiClient.post(USER_ROUTES.portfolio, payload, { auth: true }),
  updatePortfolioItem: (itemId: string, payload: Record<string, unknown>) =>
    apiClient.put(`${USER_ROUTES.portfolio}/${itemId}`, payload, { auth: true }),
  deletePortfolioItem: (itemId: string) =>
    apiClient.delete(`${USER_ROUTES.portfolio}/${itemId}`, { auth: true }),

  getWorkPreferences: () => apiClient.get(USER_ROUTES.workPreferences, { auth: true }),
  patchWorkPreferences: (payload: Record<string, unknown>) =>
    apiClient.patch(USER_ROUTES.workPreferences, payload, { auth: true }),

  getHiringPreferences: () =>
    apiClient.get(USER_ROUTES.hiringPreferences, { auth: true }),
  patchHiringPreferences: (payload: Record<string, unknown>) =>
    apiClient.patch(USER_ROUTES.hiringPreferences, payload, { auth: true }),

  listSavedFreelancers: (pageSize = 20, pageToken = "") =>
    apiClient.get(
      `${USER_ROUTES.savedFreelancers}?page_size=${pageSize}&page_token=${encodeURIComponent(pageToken)}`,
      { auth: true }
    ),
  saveFreelancer: (freelancerId: string) =>
    apiClient.post(`${USER_ROUTES.savedFreelancers}/${freelancerId}`, undefined, { auth: true }),
  removeSavedFreelancer: (freelancerId: string) =>
    apiClient.delete(`${USER_ROUTES.savedFreelancers}/${freelancerId}`, { auth: true }),

  getFreelancerNote: (freelancerId: string) =>
    apiClient.get(`${USER_ROUTES.freelancerNotes}/${freelancerId}`, { auth: true }),
  upsertFreelancerNote: (freelancerId: string, note: string) =>
    apiClient.put(`${USER_ROUTES.freelancerNotes}/${freelancerId}`, { note }, { auth: true }),
};
