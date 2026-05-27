import { baseApi } from "./baseapi";

export type NotificationType =
	| "CONNECTS_PURCHASED"
	| "CONNECTS_REFUNDED"
	| "PROPOSAL_STATUS_CHANGED"
	| "MILESTONE_STATUS_CHANGED"
	| "CONTRACT_CREATED"
	| "CONTRACT_STATUS_CHANGED"
	| "SYSTEM"
	| string;

export interface Notification {
	id: number;
	contract_id: number | null;
	created_at: string;
	is_read: boolean;
	job_id: number | null;
	message: string;
	proposal_id: number | null;
	title: string;
	type: NotificationType;
	user_id: number;
}

export interface GetNotificationsResponse {
	data: Notification[];
}

export interface MarkNotificationAsReadRequest {
	id: number;
}

export interface MarkNotificationAsReadResponse {
	message: string;
	success: boolean;
}

export const notificationApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		getNotifications: builder.query<Notification[], void>({
			query: () => "/notifications",
			transformResponse: (response: GetNotificationsResponse | Notification[]) =>
				Array.isArray(response) ? response : response.data,
		}),

		markNotificationAsRead: builder.mutation<
			MarkNotificationAsReadResponse,
			MarkNotificationAsReadRequest
		>({
			query: (body) => ({
				url: "/notifications/read",
				method: "POST",
				body,
			}),
		}),
	}),
});

export const { useGetNotificationsQuery, useMarkNotificationAsReadMutation } =
	notificationApi;