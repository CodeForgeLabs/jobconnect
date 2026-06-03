export type AvailabilityValue =
  | "AVAILABILITY_UNSPECIFIED"
  | "AVAILABILITY_FULL_TIME"
  | "AVAILABILITY_PART_TIME"
  | "AVAILABILITY_AS_NEEDED"
  | "AVAILABILITY_UNAVAILABLE";

export type ProjectLengthValue =
  | "PROJECT_LENGTH_UNSPECIFIED"
  | "PROJECT_LENGTH_SHORT_TERM"
  | "PROJECT_LENGTH_MEDIUM_TERM"
  | "PROJECT_LENGTH_LONG_TERM";

export type PortfolioMediaTypeValue =
  | "PORTFOLIO_MEDIA_TYPE_UNSPECIFIED"
  | "PORTFOLIO_MEDIA_TYPE_IMAGE"
  | "PORTFOLIO_MEDIA_TYPE_VIDEO"
  | "PORTFOLIO_MEDIA_TYPE_FILE"
  | "PORTFOLIO_MEDIA_TYPE_LINK";

export interface StringListPayload {
  values: string[];
}

export interface UserCorePayload {
  profile_id?: number;
  user_id?: string;
  role?: string;
  first_name?: string;
  last_name?: string;
  display_name?: string;
  avatar_url?: string;
  contact_email?: string;
  contact_phone?: string;
  bio?: string;
  account_status?: string;
  suspension_reason?: string;
  tax_id?: string;
  verification_status?: string;
  created_at_unix?: number;
  updated_at_unix?: number;
  location?: string;
}

export interface ClientProfilePayload {
  company_name?: string;
}

export interface FreelancerMetricsPayload {
  rating?: number;
  job_success_score?: number;
  total_reviews?: number;
  total_jobs?: number;
  total_earnings?: number;
  last_active_at_unix?: number;
}

export interface FreelancerProfilePayload {
  headline?: string;
  skills?: string[];
  hourly_rate?: number;
  availability?: AvailabilityValue | string;
  metrics?: FreelancerMetricsPayload;
}

export interface CapabilityFlagsPayload {
  can_apply_jobs?: boolean;
  can_post_jobs?: boolean;
  can_withdraw_funds?: boolean;
  can_message?: boolean;
  can_be_discovered?: boolean;
}

export interface UserProfilePayload {
  core?: UserCorePayload;
  client?: ClientProfilePayload;
  freelancer?: FreelancerProfilePayload;
  capabilities?: CapabilityFlagsPayload;
}

export interface ProfileReadinessPayload {
  percent?: number;
  missing_required_fields?: string[];
  recommendations?: string[];
}

export interface OnboardingStepPayload {
  key?: string;
  status?: string;
}

export interface GetProfileResponsePayload {
  profile?: UserProfilePayload;
  readiness?: ProfileReadinessPayload;
}

export interface PatchProfileRequestPayload {
  display_name?: string;
  contact_email?: string;
  contact_phone?: string;
  bio?: string;
  tax_id?: string;
  location?: string;
  company_name?: string;
  headline?: string;
  skills?: string[];
  hourly_rate?: number;
  availability?: AvailabilityValue | string;
}

export interface GetOnboardingStatusResponsePayload {
  readiness?: ProfileReadinessPayload;
  steps?: OnboardingStepPayload[];
}

export interface UserSettingsPayload {
  ui_locale?: string;
  email_notifications_enabled?: boolean;
  push_notifications_enabled?: boolean;
}

export interface GetSettingsResponsePayload {
  settings?: UserSettingsPayload;
}

export interface PatchSettingsRequestPayload {
  ui_locale?: string;
  email_notifications_enabled?: boolean;
  push_notifications_enabled?: boolean;
}

export interface UploadReservationRequestPayload {
  file_name: string;
  content_type: string;
}

export interface UploadReservationResponsePayload {
  storage_key: string;
  upload_url: string;
}

export interface UpsertAvatarRequestPayload {
  storage_key: string;
  file_name: string;
  content_type: string;
  width?: number;
  height?: number;
}

export interface ProfileAvatarPayload {
  user_id?: string;
  file_name?: string;
  content_type?: string;
  storage_key?: string;
  size_bytes?: number;
  width?: number;
  height?: number;
  updated_at_unix?: number;
  download_url?: string;
}

export interface UpsertAvatarResponsePayload {
  avatar_url?: string;
  avatar?: ProfileAvatarPayload;
}

export interface GetAvatarResponsePayload {
  avatar?: ProfileAvatarPayload;
}

export interface ProfileCVPayload {
  user_id?: string;
  file_name?: string;
  content_type?: string;
  size_bytes?: number;
  updated_at_unix?: number;
  download_url?: string;
}

export interface UpsertCVRequestPayload {
  storage_key: string;
  file_name: string;
  content_type: string;
}

export interface UpsertCVResponsePayload {
  cv?: ProfileCVPayload;
}

export interface GetCVResponsePayload {
  cv?: ProfileCVPayload;
}

export interface PortfolioMediaInputPayload {
  media_type: PortfolioMediaTypeValue | string;
  storage_key?: string;
  external_url?: string;
  file_name?: string;
  content_type?: string;
  size_bytes?: number;
  width?: number;
  height?: number;
}

export interface PortfolioMediaPayload extends PortfolioMediaInputPayload {
  id?: number;
  created_at_unix?: number;
}

export interface PortfolioItemPayload {
  id?: number;
  user_id?: string;
  title?: string;
  description?: string;
  project_url?: string;
  role_in_project?: string;
  completed_at_unix?: number;
  tags?: string[];
  media?: PortfolioMediaPayload[];
  created_at_unix?: number;
  updated_at_unix?: number;
}

export interface CreatePortfolioItemRequestPayload {
  title: string;
  description: string;
  project_url?: string;
  role_in_project?: string;
  completed_at_unix?: number;
  tags?: string[];
  media?: PortfolioMediaInputPayload[];
}

export interface UpdatePortfolioItemRequestPayload {
  title?: string;
  description?: string;
  project_url?: string;
  role_in_project?: string;
  completed_at_unix?: number;
  tags?: string[];
  media?: PortfolioMediaInputPayload[];
}

export interface PortfolioItemResponsePayload {
  item?: PortfolioItemPayload;
}

export interface PortfolioListResponsePayload {
  items?: PortfolioItemPayload[];
  next_page_token?: string;
}

export interface PatchWorkPreferencesRequestPayload {
  preferred_project_length?: ProjectLengthValue | string;
  min_budget?: number;
  max_budget?: number;
  contract_types?: StringListPayload;
  weekly_capacity_hours?: number;
}

export interface WorkPreferencesPayload {
  preferred_project_length?: ProjectLengthValue | string;
  min_budget?: number;
  max_budget?: number;
  contract_types?: string[];
  weekly_capacity_hours?: number;
}

export interface WorkPreferencesResponsePayload {
  settings?: WorkPreferencesPayload;
}

export interface PatchHiringPreferencesRequestPayload {
  min_hourly_rate?: number;
  max_hourly_rate?: number;
  preferred_locations?: StringListPayload;
}

export interface HiringPreferencesPayload {
  min_hourly_rate?: number;
  max_hourly_rate?: number;
  preferred_locations?: string[];
}

export interface HiringPreferencesResponsePayload {
  preferences?: HiringPreferencesPayload;
}

export interface SavedFreelancerPayload {
  freelancer_user_id?: string;
  saved_at_unix?: number;
}

export interface SavedFreelancersListResponsePayload {
  freelancers?: SavedFreelancerPayload[];
  next_page_token?: string;
}

export interface SavedFreelancerResponsePayload {
  saved?: SavedFreelancerPayload;
}

export interface FreelancerNotePayload {
  freelancer_user_id?: string;
  note?: string;
  updated_at_unix?: number;
}

export interface UpsertFreelancerNoteRequestPayload {
  note: string;
}

export interface FreelancerNoteResponsePayload {
  note?: FreelancerNotePayload;
}

export interface GenericRemovedResponse {
  removed: boolean;
}

export interface GenericDeletedResponse {
  deleted: boolean;
}

export interface VerificationRequestPayload {
  id?: number;
  user_id?: string;
  request_version?: number;
  status?: string;
  legal_name?: string;
  country_code?: string;
  document_type?: string;
  document_number_masked?: string;
  evidence_url?: string;
  submission_note?: string;
  reviewer_user_id?: string;
  rejection_reason?: string;
  internal_note?: string;
  submitted_at_unix?: number;
  reviewed_at_unix?: number;
  reverify_due_at_unix?: number;
}

export interface VerificationUploadUrlRequestPayload {
  file_name: string;
  content_type: string;
}

export interface VerificationUploadUrlResponsePayload {
  storage_key: string;
  upload_url: string;
}

export interface SubmitVerificationRequestPayload {
  legal_name: string;
  country_code: string;
  document_type: string;
  document_number_masked: string;
  evidence_url: string;
  submission_note?: string;
}

export interface SubmitVerificationResponsePayload {
  request?: VerificationRequestPayload;
}

export interface GetMyVerificationStatusResponsePayload {
  request?: VerificationRequestPayload;
}

