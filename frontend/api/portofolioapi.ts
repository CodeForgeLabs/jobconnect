// src/api/portfolioApi.ts
import { baseApi } from "./baseapi";

export interface PortfolioRequest {
  title: string;
  description: string;
  image_url: string;
  start_date: string;
  end_date: string;
  tech_stack: string[];
}

export interface PortfolioItem {
  id: number;
  user_id: number;
  title: string;
  description: string;
  image_url: string;
  start_date: string;
  end_date: string;
  tech_stack: string[];
  created_at: string;
  updated_at: string;
}

export interface PortfolioResponse {
  portfolio: PortfolioItem[];
}

export const portfolioApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    // ✅ GET portfolio
    getUserPortfolio: builder.query<PortfolioResponse, number>({
      query: (userId) => `/users/${userId}/portfolio`,
     
    }),

    // ✅ CREATE portfolio
    createPortfolio: builder.mutation<PortfolioItem, PortfolioRequest>({
      query: (body) => ({
        url: "/portfolio",
        method: "POST",
        body,
      }),
      
    }),

    // ✅ UPDATE portfolio
    updatePortfolio: builder.mutation<
      PortfolioItem,
      { id: number; data: PortfolioRequest }
    >({
      query: ({ id, data }) => ({
        url: `/portfolio/${id}`,
        method: "PATCH",
        body: data,
      }),
    
    }),

    // ✅ DELETE portfolio
    deletePortfolio: builder.mutation<unknown, number>({
      query: (id) => ({
        url: `/portfolio/${id}`,
        method: "DELETE",
      }),
     
    }),
  }),
});

export const {
  useGetUserPortfolioQuery,
  useCreatePortfolioMutation,
  useUpdatePortfolioMutation,
  useDeletePortfolioMutation,
} = portfolioApi;