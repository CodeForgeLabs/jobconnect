import { baseApi } from "./baseapi";

export type ContractStatus = "ACTIVE" | "COMPLETED" | "CANCELLED" | string;
export type ContractType = "FIXED" | "HOURLY" | string;
export type ContractMilestoneStatus =
	| "PENDING"
	| "IN_PROGRESS"
	| "SUBMITTED"
	| "REVISION_REQUESTED"
	| "APPROVED"
	| "PAID"
	| string;

export type MilestoneStatus = ContractMilestoneStatus;

export interface CreateContractRequest {
	freelancer_id: string;
	job_id: string;
}

export interface SubmitMilestoneWorkRequest {
	contract_id: number;
	description: string;
	milestone_id: number;
	milestone_project_url: string;
}

export interface WorkSessionRequest {
	contract_id: number;
}

export interface PayWeeklyLogsRequest {
	contract_id: number;
	week_number: number;
	year: number;
}

export interface WorkSessionLogEntry {
	id: number;
	start_time: string;
	end_time: string;
	total_hours: number;
	is_paid: boolean;
}

export interface WorkSessionTimeLogApiEntry {
	ID: number;
	ContractID: number;
	FreelancerID: number;
	start_time: string;
	end_time?: string;
	TotalHours: number;
	IsPaid: boolean;
	CreatedAt: string;
	UpdatedAt: string;
}

export interface WorkSessionTimeLogsResponse {
	time_logs: WorkSessionTimeLogApiEntry[];
}

export interface WorkSessionDayLog {
	day: string;
	date: string;
	total_hours: number;
	sessions: WorkSessionLogEntry[];
}

export interface WeeklyWorkLog {
	week_number: number;
	week_start: string;
	week_end: string;
	total_hours: number;
	days: WorkSessionDayLog[];
}

export interface WeeklyWorkLogsResponse {
	data: WeeklyWorkLog[];
}

export interface UpdateContractStatusRequest {
	contractId: number;
	newStatus: ContractStatus;

}

export interface UpdateMilestoneStatusRequest {
	milestoneId: number;
	newStatus: ContractMilestoneStatus;
	feedback?: string;
	
}

export interface ContractMilestone {
	Amount: number;
	ClientFeedback: string;
	ContractID: number;
	CreatedAt: string;
	Description: string;
	Due_date: string;
	ID: number;
	Status: ContractMilestoneStatus;
	submission_url: string;
	UpdatedAt: string;
	WorkDescription: string;
}

export interface Contract {
	client_email: string;
	client_first_name: string;
	client_headline: string;
	client_id: number;
	client_last_name: string;
	client_profile_picture_url: string;
	client_skills: string;
	contract_id: number;
	description: string;
	end_date: string;
	freelancer_email: string;
	freelancer_first_name: string;
	freelancer_headline: string;
	freelancer_id: number;
	freelancer_last_name: string;
	freelancer_profile_picture_url: string;
	freelancer_skills: string;
	hourly_rate: number;
	job_id: number;
	job_title: string;
	milestones: ContractMilestone[];
	proposal_description: string;
	start_date: string;
	status: ContractStatus;
	title: string;
	total_budget: number;
	type: ContractType;
	weekly_hour_limit: number;
}

export interface ApiMessageResponse {
	message: string;
}

const normalizeContractsList = (response: unknown): Contract[] => {
	if (Array.isArray(response)) {
		return response as Contract[];
	}

	if (!response || typeof response !== "object") {
		return [];
	}

	const maybeContracts = (response as { contracts?: unknown }).contracts;
	if (Array.isArray(maybeContracts)) {
		return maybeContracts as Contract[];
	}

	const maybeData = (response as { data?: unknown }).data;
	if (Array.isArray(maybeData)) {
		return maybeData as Contract[];
	}

	return [];
};

const normalizeContract = (response: unknown): Contract | null => {
	if (!response || typeof response !== "object") {
		return null;
	}

	if ("contract" in response && (response as { contract?: unknown }).contract) {
		return (response as { contract?: Contract }).contract ?? null;
	}

	if ("data" in response && (response as { data?: unknown }).data) {
		return (response as { data?: Contract }).data ?? null;
	}

	if ("contract_id" in response) {
		return response as Contract;
	}

	return null;
};

export const contractApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		createContract: builder.mutation<ApiMessageResponse, CreateContractRequest>({
			query: (body) => ({
				url: "/contracts",
				method: "POST",
				body,
			}),
		}),

		submitMilestoneWork: builder.mutation<
			ApiMessageResponse,
			SubmitMilestoneWorkRequest
		>({
			query: (body) => ({
				url: "/contracts/milestone/submit",
				method: "POST",
				body,
			}),
		}),

		updateMilestoneStatus: builder.mutation<
			ApiMessageResponse,
			UpdateMilestoneStatusRequest
		>({
			query: ({ milestoneId, newStatus , feedback }) => ({
				url: `/contracts/milestone/${milestoneId}/status?new_status=${encodeURIComponent(newStatus)}`,
				method: "PATCH",
				body: {
					feedback
				}
			}),
		}),

		getMyContracts: builder.query<Contract[], void>({
			query: () => "/contracts/mine",
			transformResponse: (response: unknown) => normalizeContractsList(response),
		}),

		startWorkSession: builder.mutation<ApiMessageResponse, WorkSessionRequest>({
			query: (body) => ({
				url: "/contracts/work-session/start",
				method: "POST",
				body,
			}),
		}),

		endWorkSession: builder.mutation<ApiMessageResponse, WorkSessionRequest>({
			query: (body) => ({
				url: "/contracts/work-session/end",
				method: "POST",
				body,
			}),
		}),

		getWorkSessionTimeElapsed: builder.mutation<Record<string, unknown>, WorkSessionRequest>({
			query: (body) => ({
				url: "/contracts/work-session/time-elapsed",
				method: "POST",
				body,
			}),
		}),

		getWorkSessionTimeLogs: builder.mutation<WorkSessionTimeLogsResponse, WorkSessionRequest>({
			query: (body) => ({
				url: "/contracts/work-session/time-logs",
				method: "POST",
				body,
			}),
		}),

		getWeeklyHours: builder.mutation<Record<string, unknown>, WorkSessionRequest>({
			query: (body) => ({
				url: "/contracts/work-session/weekly-hours",
				method: "POST",
				body,
			}),
		}),

		getWeeklyLogs: builder.mutation<WeeklyWorkLogsResponse, WorkSessionRequest>({
			query: (body) => ({
				url: "/contracts/work-session/weekly-logs",
				method: "POST",
				body,
			}),
		}),

		payWeeklyLogs: builder.mutation<ApiMessageResponse, PayWeeklyLogsRequest>({
			query: (body) => ({
				url: "/contracts/work-session/pay-weekly-logs",
				method: "POST",
				body,
			}),
		}),

		updateContractStatus: builder.mutation<
			ApiMessageResponse,
			UpdateContractStatusRequest
		>({
			query: ({ contractId, newStatus }) => ({
				url: `/contracts/${contractId}/status?new_status=${encodeURIComponent(newStatus)}`,
				method: "PATCH",
			}),
		}),

		getContractById: builder.query<Contract, number>({
			query: (id) => `/contracts/${id}`,
			transformResponse: (response: unknown) =>
				normalizeContract(response) ?? (response as Contract),
		}),
	}),
});

export const {
	useCreateContractMutation,
	useSubmitMilestoneWorkMutation,
	useUpdateMilestoneStatusMutation,
	useGetMyContractsQuery,
	useStartWorkSessionMutation,
	useEndWorkSessionMutation,
	useGetWorkSessionTimeElapsedMutation,
	useGetWorkSessionTimeLogsMutation,
	useGetWeeklyHoursMutation,
	useGetWeeklyLogsMutation,
	usePayWeeklyLogsMutation,
	useUpdateContractStatusMutation,
	useGetContractByIdQuery,
} = contractApi;
