
import { baseApi } from "./baseapi";
import type { FetchBaseQueryError } from '@reduxjs/toolkit/query';


export type Availability = "FULLTIME" | "PARTTIME" | "FREELANCE";

export type Role = "CLIENT" | "FREELANCER";

export interface User {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  bio: string;
  headline: string;
  location: string;
  phone_number: string;
  profile_picture_url: string;
  skills: string | string[];
  hourly_rate: number;
  availability: Availability;
  connect: number;
  role: Role;
  created_at: string;
  updated_at: string;
}

/* ========================
   🔐 REQUEST TYPES
======================== */

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  first_name: string;
  last_name: string;
  password: string;
  company_name?: string;
  role: "CLIENT" | "FREELANCER";
}

export interface FetchUsersParams {
  skills?: string;
  location?: string;
  min_hourly_rate?: number;
}

export interface SendOtpRequest {
  email: string;
}

export interface VerifyOtpRequest {
  email: string;
  otp: string;
}

/* ========================
   🌐 API
======================== */

export const userApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({

    // 🔐 LOGIN (sets HTTP-only cookie)
    login: builder.mutation<unknown, LoginRequest>({
      query: (body) => ({
        url: "/users/login",
        method: "POST",
        body,
      }),
    }),

    logout: builder.mutation({
      query: () => ({
        url: "/users/logout",
        method: "POST",
      }),
    }),

    // 🆕 REGISTER
    register: builder.mutation<unknown, RegisterRequest>({
      query: (body) => ({
        url: "/users/register",
        method: "POST",
        body,
      }),
    }),

    // 👤 CURRENT USER
    getMe: builder.query<User, void>({
      query: () => "/users/me",
   
    }),

    // 🔎 GET USER BY EMAIL
    getUserByEmail: builder.query<User, string>({
      query: (email) => `/users/email?email=${email}`,
    }),

    // 🔎 GET USER BY ID
    getUserById: builder.query<User, number>({
      query: (id) => `/users/byid?id=${id}`,
    }),

    // 🔎 SEARCH USERS BY NAME
    searchUsersByName: builder.query<User[], string>({
      query: (name) => `/users/search?name=${name}`,
    }),

    // 🔎 LIST USERS WITH FILTERS
    fetchUsers: builder.query<User[], FetchUsersParams | void>({
      query: (filters) => ({
        url: "/users/fetch",
        params: {
          ...(filters?.skills ? { skills: filters.skills } : {}),
          ...(filters?.location ? { location: filters.location } : {}),
          ...(filters?.min_hourly_rate !== undefined
            ? { min_hourly_rate: filters.min_hourly_rate }
            : {}),
        },
      }),
    }),

    // 📩 SEND OTP
    sendOtp: builder.mutation<Record<string, string>, SendOtpRequest>({
      query: ({ email }) => ({
        url: "/users/send-otp",
        method: "POST",
        params: { email },
      }),
    }),

    // ✅ VERIFY OTP
    verifyOtp: builder.mutation<Record<string, string>, VerifyOtpRequest>({
      query: ({ email, otp }) => ({
        url: "/users/verify-otp",
        method: "POST",
        params: { email, otp },
      }),
    }),

    // ✏️ UPDATE USER (Cloudinary URL goes here)
    updateMe: builder.mutation<User, Partial<User>>({
      query: (body) => ({
        url: "/users/me",
        method: "PATCH",
        body,
      }),
   
    }),

    // 📤 UPLOAD IMAGE
    uploadImage: builder.mutation<{ secure_url: string }, File>({
      queryFn: async (file) => {
        try {
      const formData = new FormData();

      formData.append("file", file);
      formData.append(
        "upload_preset",
        process.env.NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET!
      );

      const res = await fetch(
        `https://api.cloudinary.com/v1_1/${process.env.NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME}/image/upload`,
        {
          method: "POST",
          body: formData,
        }
      );

      const data = await res.json();

      return { data: { secure_url: String(data.secure_url) } };
        } catch (error) {
          return { error: error as FetchBaseQueryError };
        }
      },
    }),


    uploadFile: builder.mutation<{ secure_url: string }, File>({
      queryFn: async (file) => {
        try {
      const formData = new FormData();

      formData.append("file", file);
      formData.append(
        "upload_preset",
        process.env.NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET!
      );

      const res = await fetch(
        `https://api.cloudinary.com/v1_1/${process.env.NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME}/raw/upload`,
        {
          method: "POST",
          body: formData,
        }
      );

      const data = await res.json();

      return { data: { secure_url: String(data.secure_url) } };
        } catch (error) {
          return { error: error as FetchBaseQueryError };
        }
      },
    }),

    // ❌ DELETE ACCOUNT
    deleteMe: builder.mutation<void, void>({
      query: () => ({
        url: "/users/me",
        method: "DELETE",
      }),
    }),
  }),
});

export const {
  useLoginMutation,
  useLogoutMutation,
  useRegisterMutation,
  useGetMeQuery,
  useGetUserByEmailQuery,
  useGetUserByIdQuery,
  useSearchUsersByNameQuery,
  useFetchUsersQuery,
  useSendOtpMutation,
  useVerifyOtpMutation,
  useUpdateMeMutation,
  useDeleteMeMutation,
  useUploadImageMutation,
  useUploadFileMutation,
} = userApi;