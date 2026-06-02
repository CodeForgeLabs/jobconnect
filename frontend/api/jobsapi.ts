import { baseApi } from "./baseapi";

export type JobStatus = "OPEN" | "CLOSED" | "PAUSED";
export type JobType = "HOURLY" | "FIXED";
export type WorkMode = "REMOTE" | "ONSITE" | "HYBRID";
export type ExperienceLevel = "ENTRY" | "INTERMEDIATE" | "EXPERT";

export interface JobMilestone {
	id?: number;
	amount: number;
	description: string;
	created_at?: string;
	deadline?: string;
	is_paid?: boolean;
	job_id?: number;
}

export interface Job {
	id: number;
	applications_count: number;
	budget: number;
	category: string;
	company_name: string;
	created_at: Date | string;
	created_by: number;
	deadline: string;
	description: string;
	experience_level: ExperienceLevel | string;
	hourly_rate: number;
	is_private: boolean;
	job_type: JobType | string;
	location: string;
	max_weekly_hours: number;
	milestones: JobMilestone[];
	skills: string;
	status: JobStatus | string;
	title: string;
	updated_at: Date | string;
	work_mode: WorkMode | string;
}

export interface JobFilterParams {
	title?: string;
	category?: string;
	job_type?: string;
	work_mode?: string;
	experience_level?: string;
	budget_min?: number;
}

export interface CreateJobRequest {
	budget: number;
	category: string;
	company_name: string;
	description: string;
	experience_level: string;
	hourly_rate: number;
	is_private: boolean;
	job_type: "FIXED" | "HOURLY";
	location: string;
	max_weekly_hours: number;
	milestones: Array<{
		amount: number;
		description: string;
		deadline: Date;
	}>;
	skills: string[];
	title: string;
	work_mode: string;
}

export interface InviteUserRequest {
  job_id: number;
  user_id: number;
}

export interface UpdateJobRequest {
	budget?: number;
	category?: string;
	company_name?: string;
	description?: string;
	experience_level?: string;
	hourly_rate?: number;
	is_private?: boolean;
	job_type?: "FIXED" | "HOURLY";
	location?: string;
	max_weekly_hours?: number;
	skills?: string[];
	status?: string;
	title?: string;
	work_mode?: string;
}

const normalizeJobsListResponse = (response: unknown): Job[] => {
	if (Array.isArray(response)) {
		return response as Job[];
	}

	if (response && typeof response === "object") {
		const maybeJobs = (response as { jobs?: unknown }).jobs;
		if (Array.isArray(maybeJobs)) {
			return maybeJobs as Job[];
		}

		const maybeData = (response as { data?: unknown }).data;
		if (Array.isArray(maybeData)) {
			return maybeData as Job[];
		}
	}

	return [];
};

export const jobsApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		getJobs: builder.query<Job[], JobFilterParams | void>({
			query: (filters) => {
				const searchParams = new URLSearchParams();

				if (filters) {
					if (filters.title) searchParams.set("title", filters.title);
					if (filters.category) searchParams.set("category", filters.category);
					if (filters.job_type) searchParams.set("job_type", filters.job_type);
					if (filters.work_mode) searchParams.set("work_mode", filters.work_mode);
					if (filters.experience_level) {
						searchParams.set("experience_level", filters.experience_level);
					}
					if (filters.budget_min !== undefined) {
						searchParams.set("budget_min", String(filters.budget_min));
					}
				}

				const queryString = searchParams.toString();

				return queryString ? `/jobs?${queryString}` : "/jobs";
			},
			transformResponse: (response: unknown) =>
				normalizeJobsListResponse(response),
		}),

		getJobsRecommended: builder.query<Job[], JobFilterParams | void>({
			query: (filters) => {
				const searchParams = new URLSearchParams();

				if (filters) {
					if (filters.title) searchParams.set("title", filters.title);
					if (filters.category) searchParams.set("category", filters.category);
					if (filters.job_type) searchParams.set("job_type", filters.job_type);
					if (filters.work_mode) searchParams.set("work_mode", filters.work_mode);
					if (filters.experience_level) {
						searchParams.set("experience_level", filters.experience_level);
					}
					if (filters.budget_min !== undefined) {
						searchParams.set("budget_min", String(filters.budget_min));
					}
				}

				const queryString = searchParams.toString();

				return queryString ? `/jobs/fetch/recommended?${queryString}` : "jobs/fetch/recommended";
			},
			transformResponse: (response: unknown) =>
				normalizeJobsListResponse(response),
		}),

		createJob: builder.mutation<Job, CreateJobRequest>({
			query: (body) => ({
				url: "/jobs",
				method: "POST",
				body,
			}),
		}),

		getMyJobs: builder.query<Job[], void>({
			query: () => "/jobs/mine",
			transformResponse: (response: unknown) =>
				normalizeJobsListResponse(response),
		}),
		getInvitedJobs: builder.query<Job[], void>({
			query: () => "/jobs/fetch/invited",
			transformResponse: (response: unknown) =>
				normalizeJobsListResponse(response),
		}),


		getJobsByClient: builder.query<Job[], number>({
            query: (clientId) => ({
                url: "/jobs/by-client",
                method: "GET",
                params: { id: clientId },
            }),
            transformResponse: (response: unknown) =>
                normalizeJobsListResponse(response),
        }),

		getJobById: builder.query<{ job: Job }, number>({
			query: (id) => `/jobs/${id}`,
		}),

		deleteJob: builder.mutation<Record<string, string>, number>({
			query: (id) => ({
				url: `/jobs/${id}`,
				method: "DELETE",
			}),
		}),

		inviteToJob: builder.mutation<void, InviteUserRequest>({
				query: (body) => ({
					url: "/jobs/invite",
					method: "POST",
					body,
				}),
				}),

		updateJob: builder.mutation<Job, { id: number; body: UpdateJobRequest }>({
			query: ({ id, body }) => ({
				url: `/jobs/${id}`,
				method: "PATCH",
				body,
			}),
		}),
	}),
});

export const {
	useGetJobsQuery,
	useCreateJobMutation,
	useGetMyJobsQuery,
	useGetJobByIdQuery,
	useDeleteJobMutation,
	useUpdateJobMutation,
	useInviteToJobMutation,
    useGetJobsRecommendedQuery,
	useGetJobsByClientQuery,
	useGetInvitedJobsQuery,
} = jobsApi;
