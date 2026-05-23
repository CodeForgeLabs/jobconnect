import { baseApi } from "./baseapi";

export type ProposalStatus = "PENDING" | "APPROVED" | "REJECTED" | string;

export interface CreateProposalRequest {
	cover_letter: string;
	job_id: number;
}

export interface UpdateProposalRequest {
	cover_letter?: string;
	status?: ProposalStatus;
}

export interface Proposal {
	description?: string;
	id: number;
	job_id: number;
	sender_id: number;
	status: ProposalStatus;
    created_at?: string;
    updated_at?: string;
    job_owner_id?: number;
  
          
}

export interface ProposalApplicant {
	description?: string;
	email?: string;
	first_name?: string;
	headline?: string;
	job_id: number;
	last_name?: string;
	profile_picture_url?: string;
	proposal_id?: number;
	sender_id?: number;
	skills?: string;
	status?: ProposalStatus;
	user_id?: number;
}

export interface ProposalListResponse {
	proposals: Proposal[];
}

const normalizeProposal = (response: unknown): Proposal | null => {
	if (!response || typeof response !== "object") return null;

	const candidate =
		(response as { proposal?: unknown }).proposal ??
		(response as { data?: unknown }).data ??
		response;

	if (!candidate || typeof candidate !== "object") return null;

	if (!("id" in candidate) || !("job_id" in candidate)) return null;

	return candidate as Proposal;
};

const normalizeProposalsList = (response: unknown): Proposal[] => {
	if (Array.isArray(response)) return response as Proposal[];

	if (!response || typeof response !== "object") return [];

	const maybeProposals = (response as { proposals?: unknown }).proposals;
	if (Array.isArray(maybeProposals)) return maybeProposals as Proposal[];

	const maybeData = (response as { data?: unknown }).data;
	if (Array.isArray(maybeData)) return maybeData as Proposal[];

	return [];
};

const normalizeProposalApplicants = (response: unknown): ProposalApplicant[] => {
	if (response && typeof response === "object" && "proposals" in response) {
		return (response as { proposals?: unknown }).proposals as ProposalApplicant[];
	}

	if (!response || typeof response !== "object") return [];

	const maybeApplicants = (response as { applicants?: unknown }).applicants;
	if (Array.isArray(maybeApplicants)) return maybeApplicants as ProposalApplicant[];

	const maybeData = (response as { data?: unknown }).data;
	if (Array.isArray(maybeData)) return maybeData as ProposalApplicant[];

	return [];
};

export const proposalApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		createProposal: builder.mutation<Proposal, CreateProposalRequest>({
			query: (body) => ({
				url: "/proposals",
				method: "POST",
				body,
			}),
			transformResponse: (response: unknown) => {
				const normalized = normalizeProposal(response);
				return normalized ?? ({ id: 0, job_id: 0, sender_id: 0, status: "PENDING" } as Proposal);
			},
		}),

		getMyProposals: builder.query<Proposal[], void>({
			query: () => "/proposals/mine",
			transformResponse: (response: unknown) => normalizeProposalsList(response),
		}),

		getProposalById: builder.query<Proposal, number>({
			query: (id) => `/proposals/${id}`,
			transformResponse: (response: unknown) => {
				const normalized = normalizeProposal(response);
				return normalized ?? ({ id: 0, job_id: 0, sender_id: 0, status: "PENDING" } as Proposal);
			},
		}),

		getJobProposals: builder.mutation<ProposalApplicant[], { job_id: number }>({
			query: (body) => ({
				url: "/proposals/jobs",
				method: "POST",
				body,
			}),
			transformResponse: (response: unknown) =>
				normalizeProposalApplicants(response),
		}),

		updateProposal: builder.mutation<Proposal, { id: number; payload: UpdateProposalRequest }>({
			query: ({ id, payload }) => ({
				url: `/proposals/${id}`,
				method: "PATCH",
				body: payload,
			}),
			transformResponse: (response: unknown) => {
				const normalized = normalizeProposal(response);
				return normalized ?? ({ id: 0, job_id: 0, sender_id: 0, status: "pending" } as Proposal);
			},
		}),

		deleteProposal: builder.mutation<Record<string, string>, number>({
			query: (id) => ({
				url: `/proposals/${id}`,
				method: "DELETE",
			}),
		}),
	}),
});

export const {
	useCreateProposalMutation,
	useGetMyProposalsQuery,
	useGetProposalByIdQuery,
	useGetJobProposalsMutation,
	useUpdateProposalMutation,
	useDeleteProposalMutation,
} = proposalApi;
