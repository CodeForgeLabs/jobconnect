import { baseApi } from "./baseapi";

export interface Review {
	id: number;
	client_id: number;
	freelancer_id: number;
	contract_id: number;
	note: string;
	rating: number;
	freelancer_reply?: string | null;
	created_at: string;
	updated_at: string;
}

export interface CreateReviewRequest {
	contract_id: number;
	note: string;
	rating: number;
}

export interface UpdateReviewRequest {
	note?: string;
	rating?: number;
}

export interface UpdateReplyRequest {
	reply: string;
}

export interface UserReviewsResponse {
	average_rating: number;
	review_count: number;
	reviews: Review[];
}

export const reviewsApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		createReview: builder.mutation<Review, CreateReviewRequest>({
			query: (body) => ({
				url: "/reviews",
				method: "POST",
				body,
			}),
		}),

		updateReview: builder.mutation<Review, { id: number; body: UpdateReviewRequest }>({
			query: ({ id, body }) => ({
				url: `/reviews/${id}`,
				method: "PATCH",
				body,
			}),
		}),

		updateReviewReply: builder.mutation<Review, { id: number; body: UpdateReplyRequest }>({
			query: ({ id, body }) => ({
				url: `/reviews/${id}/reply`,
				method: "PATCH",
				body,
			}),
		}),

		getUserReviews: builder.query<UserReviewsResponse, number>({
			query: (userId) => `/users/${userId}/reviews`,
		}),
	}),
});

export const {
	useCreateReviewMutation,
	useUpdateReviewMutation,
	useUpdateReviewReplyMutation,
	useGetUserReviewsQuery,
} = reviewsApi;

export default reviewsApi;

