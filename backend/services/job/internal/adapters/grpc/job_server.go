package grpcadapter

import (
	"context"
	"strings"

	jobv1 "jobconnect/job/gen/job/v1"
	"jobconnect/job/internal/application"
	"jobconnect/job/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JobServer struct {
	jobv1.UnimplementedJobServiceServer
	CreateJobUC        *application.CreateJob
	GetJobUC           *application.GetJob
	GetJobSummaryUC    *application.GetJobSummary
	UpdateJobUC        *application.UpdateJob
	ListMyJobsUC       *application.ListMyJobs
	ListOpenJobsUC     *application.ListOpenJobs
	CloseJobUC         *application.CloseJob
	UploadAttachmentUC *application.UploadJobAttachment
	DeleteAttachmentUC *application.DeleteJobAttachment
	InviteFreelancerUC *application.InviteFreelancerToJob
	ListApplicantsUC   *application.ListJobApplicants
	SetApplicantUC     *application.SetApplicantStage
	SetVisibilityUC    *application.SetJobVisibility
	SetBudgetRangeUC   *application.SetJobBudgetRange
	PauseJobUC         *application.PauseJob
	ReopenJobUC        *application.ReopenJob
	MarkFilledUC       *application.MarkJobFilled
	SearchJobsUC       *application.SearchJobs
	ListFacetsUC       *application.ListJobFacets
	ListAttachmentsUC  *application.ListJobAttachments
	GetAttachmentURLUC *application.GetJobAttachmentDownloadURL
	GetPublicJobUC     *application.GetPublicJobDetail
	ListInvitedJobsUC  *application.ListInvitedJobs
	RespondInviteUC    *application.RespondToJobInvite
	SaveJobUC          *application.SaveJob
	UnsaveJobUC        *application.UnsaveJob
	ListSavedJobsUC    *application.ListSavedJobs
	HireApplicantUC    *application.HireApplicant
	RejectAllUC        *application.RejectAllApplicants
	ReopenHiringUC     *application.ReopenHiringForJob
	GetJobStatsUC      *application.GetJobStats
	SearchJobsV2UC     *application.SearchJobsV2
	MarkCompletedUC    *application.MarkJobCompleted
	CancelWithSettleUC *application.CancelJobWithSettlementPolicy
	TokenParser        TokenParser
}

type JobServerConfig struct {
	CreateJobUC        *application.CreateJob
	GetJobUC           *application.GetJob
	GetJobSummaryUC    *application.GetJobSummary
	UpdateJobUC        *application.UpdateJob
	ListMyJobsUC       *application.ListMyJobs
	ListOpenJobsUC     *application.ListOpenJobs
	CloseJobUC         *application.CloseJob
	UploadAttachmentUC *application.UploadJobAttachment
	DeleteAttachmentUC *application.DeleteJobAttachment
	InviteFreelancerUC *application.InviteFreelancerToJob
	ListApplicantsUC   *application.ListJobApplicants
	SetApplicantUC     *application.SetApplicantStage
	SetVisibilityUC    *application.SetJobVisibility
	SetBudgetRangeUC   *application.SetJobBudgetRange
	PauseJobUC         *application.PauseJob
	ReopenJobUC        *application.ReopenJob
	MarkFilledUC       *application.MarkJobFilled
	SearchJobsUC       *application.SearchJobs
	ListFacetsUC       *application.ListJobFacets
	ListAttachmentsUC  *application.ListJobAttachments
	GetAttachmentURLUC *application.GetJobAttachmentDownloadURL
	GetPublicJobUC     *application.GetPublicJobDetail
	ListInvitedJobsUC  *application.ListInvitedJobs
	RespondInviteUC    *application.RespondToJobInvite
	SaveJobUC          *application.SaveJob
	UnsaveJobUC        *application.UnsaveJob
	ListSavedJobsUC    *application.ListSavedJobs
	HireApplicantUC    *application.HireApplicant
	RejectAllUC        *application.RejectAllApplicants
	ReopenHiringUC     *application.ReopenHiringForJob
	GetJobStatsUC      *application.GetJobStats
	SearchJobsV2UC     *application.SearchJobsV2
	MarkCompletedUC    *application.MarkJobCompleted
	CancelWithSettleUC *application.CancelJobWithSettlementPolicy
	TokenParser        TokenParser
}

func NewJobServer(cfg JobServerConfig) *JobServer {
	return &JobServer{
		CreateJobUC:        cfg.CreateJobUC,
		GetJobUC:           cfg.GetJobUC,
		GetJobSummaryUC:    cfg.GetJobSummaryUC,
		UpdateJobUC:        cfg.UpdateJobUC,
		ListMyJobsUC:       cfg.ListMyJobsUC,
		ListOpenJobsUC:     cfg.ListOpenJobsUC,
		CloseJobUC:         cfg.CloseJobUC,
		UploadAttachmentUC: cfg.UploadAttachmentUC,
		DeleteAttachmentUC: cfg.DeleteAttachmentUC,
		InviteFreelancerUC: cfg.InviteFreelancerUC,
		ListApplicantsUC:   cfg.ListApplicantsUC,
		SetApplicantUC:     cfg.SetApplicantUC,
		SetVisibilityUC:    cfg.SetVisibilityUC,
		SetBudgetRangeUC:   cfg.SetBudgetRangeUC,
		PauseJobUC:         cfg.PauseJobUC,
		ReopenJobUC:        cfg.ReopenJobUC,
		MarkFilledUC:       cfg.MarkFilledUC,
		SearchJobsUC:       cfg.SearchJobsUC,
		ListFacetsUC:       cfg.ListFacetsUC,
		ListAttachmentsUC:  cfg.ListAttachmentsUC,
		GetAttachmentURLUC: cfg.GetAttachmentURLUC,
		GetPublicJobUC:     cfg.GetPublicJobUC,
		ListInvitedJobsUC:  cfg.ListInvitedJobsUC,
		RespondInviteUC:    cfg.RespondInviteUC,
		SaveJobUC:          cfg.SaveJobUC,
		UnsaveJobUC:        cfg.UnsaveJobUC,
		ListSavedJobsUC:    cfg.ListSavedJobsUC,
		HireApplicantUC:    cfg.HireApplicantUC,
		RejectAllUC:        cfg.RejectAllUC,
		ReopenHiringUC:     cfg.ReopenHiringUC,
		GetJobStatsUC:      cfg.GetJobStatsUC,
		SearchJobsV2UC:     cfg.SearchJobsV2UC,
		MarkCompletedUC:    cfg.MarkCompletedUC,
		CancelWithSettleUC: cfg.CancelWithSettleUC,
		TokenParser:        cfg.TokenParser,
	}
}

func (s *JobServer) CreateJob(ctx context.Context, req *jobv1.CreateJobRequest) (*jobv1.CreateJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}

	var deadline *int64
	if req.DeadlineUnixSeconds > 0 {
		deadline = &req.DeadlineUnixSeconds
	}

	attachments := make([]domain.Attachment, 0, len(req.Attachments))
	for _, a := range req.Attachments {
		if a == nil {
			continue
		}
		attachments = append(attachments, domain.Attachment{
			FileName:    a.FileName,
			ContentType: a.ContentType,
			URL:         a.Url,
			SizeBytes:   a.SizeBytes,
		})
	}
	jobType, mapErr := jobTypeFromEnum(req.JobTypeEnum)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}

	out, err := s.CreateJobUC.Execute(ctx, application.CreateJobInput{
		ClientID:       callerID,
		Title:          req.Title,
		Description:    req.Description,
		RequiredSkills: req.RequiredSkills,
		JobType:        jobType,
		BudgetFixed:    req.BudgetFixed,
		HourlyRate:     req.HourlyRate,
		Currency:       req.Currency,
		Deadline:       deadline,
		Attachments:    attachments,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &jobv1.CreateJobResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) GetJob(ctx context.Context, req *jobv1.GetJobRequest) (*jobv1.GetJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.GetJobUC.Execute(ctx, application.GetJobInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.GetJobResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) GetJobSummary(ctx context.Context, req *jobv1.GetJobSummaryRequest) (*jobv1.GetJobSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "client" && role != "freelancer" {
		return nil, status.Error(codes.PermissionDenied, "client or freelancer role required")
	}

	out, err := s.GetJobSummaryUC.Execute(ctx, application.GetJobSummaryInput{JobID: req.JobId})
	if err != nil {
		return nil, toStatus(err)
	}

	resp := &jobv1.GetJobSummaryResponse{}
	if !out.Summary.Found {
		resp.Summary = &jobv1.JobSummary{JobId: req.JobId, Found: false}
		return resp, nil
	}

	resp.Summary = &jobv1.JobSummary{
		JobId:    out.Summary.JobID,
		ClientId: out.Summary.ClientID,
		Status:   jobStatusToEnum(out.Summary.Status),
		IsOpen:   out.Summary.IsOpen,
		Found:    out.Summary.Found,
	}
	return resp, nil
}

func (s *JobServer) UpdateJob(ctx context.Context, req *jobv1.UpdateJobRequest) (*jobv1.UpdateJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}

	in := application.UpdateJobInput{
		JobID:               req.JobId,
		ClientID:            callerID,
		ClearDeadline:       req.ClearDeadline,
		ClearRequiredSkills: req.ClearRequiredSkills,
		ClearAttachments:    req.ClearAttachments,
	}
	if req.Title != nil {
		in.Title = req.Title
	}
	if req.Description != nil {
		in.Description = req.Description
	}
	if len(req.RequiredSkills) > 0 {
		in.RequiredSkills = req.RequiredSkills
	}
	if req.JobTypeEnum != nil {
		mapped, mapErr := jobTypeFromEnum(req.GetJobTypeEnum())
		if mapErr != nil {
			return nil, status.Error(codes.InvalidArgument, mapErr.Error())
		}
		in.JobType = &mapped
	}
	if req.BudgetFixed != nil {
		in.BudgetFixed = req.BudgetFixed
	}
	if req.HourlyRate != nil {
		in.HourlyRate = req.HourlyRate
	}
	if req.Currency != nil {
		in.Currency = req.Currency
	}
	if req.DeadlineUnixSeconds != nil {
		in.Deadline = req.DeadlineUnixSeconds
	}
	if len(req.Attachments) > 0 {
		attachments := make([]domain.Attachment, 0, len(req.Attachments))
		for _, a := range req.Attachments {
			if a == nil {
				continue
			}
			attachments = append(attachments, domain.Attachment{
				FileName:    a.FileName,
				ContentType: a.ContentType,
				URL:         a.Url,
				SizeBytes:   a.SizeBytes,
			})
		}
		in.Attachments = attachments
	}

	out, err := s.UpdateJobUC.Execute(ctx, in)
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.UpdateJobResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) ListMyJobs(ctx context.Context, req *jobv1.ListMyJobsRequest) (*jobv1.ListMyJobsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.ListMyJobsUC.Execute(ctx, application.ListMyJobsInput{
		ClientID:  callerID,
		Status:    jobStatusFromEnum(req.StatusEnum),
		PageSize:  req.PageSize,
		PageToken: req.PageToken,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	jobs := make([]*jobv1.Job, 0, len(out.Jobs))
	for _, j := range out.Jobs {
		jobs = append(jobs, toProtoJob(j))
	}
	return &jobv1.ListMyJobsResponse{Jobs: jobs, NextPageToken: out.NextPageToken}, nil
}

func (s *JobServer) ListOpenJobs(ctx context.Context, req *jobv1.ListOpenJobsRequest) (*jobv1.ListOpenJobsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	jobTypeFilter := ""
	if req.JobTypeEnum != jobv1.JobType_JOB_TYPE_UNSPECIFIED {
		mapped, mapErr := jobTypeFromEnum(req.JobTypeEnum)
		if mapErr != nil {
			return nil, status.Error(codes.InvalidArgument, mapErr.Error())
		}
		jobTypeFilter = mapped
	}

	out, err := s.ListOpenJobsUC.Execute(ctx, application.ListOpenJobsInput{
		PageSize:    req.PageSize,
		PageToken:   req.PageToken,
		SearchQuery: req.SearchQuery,
		Skills:      req.Skills,
		JobType:     jobTypeFilter,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	jobs := make([]*jobv1.Job, 0, len(out.Jobs))
	for _, j := range out.Jobs {
		jobs = append(jobs, toProtoJob(j))
	}
	return &jobv1.ListOpenJobsResponse{Jobs: jobs, NextPageToken: out.NextPageToken}, nil
}

func (s *JobServer) CloseJob(ctx context.Context, req *jobv1.CloseJobRequest) (*jobv1.CloseJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	reason, mapErr := closeReasonFromEnum(req.ReasonEnum)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}

	out, err := s.CloseJobUC.Execute(ctx, application.CloseJobInput{JobID: req.JobId, ClientID: callerID, Reason: reason})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.CloseJobResponse{Closed: out.Closed}, nil
}

func (s *JobServer) UploadJobAttachment(ctx context.Context, req *jobv1.UploadJobAttachmentRequest) (*jobv1.UploadJobAttachmentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}

	out, err := s.UploadAttachmentUC.Execute(ctx, application.UploadJobAttachmentInput{
		JobID:        req.JobId,
		ClientID:     callerID,
		FileName:     req.FileName,
		ContentType:  req.ContentType,
		ContentBytes: req.Content,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &jobv1.UploadJobAttachmentResponse{Attachment: &jobv1.JobAttachment{
		Id:          out.Attachment.ID,
		FileName:    out.Attachment.FileName,
		ContentType: out.Attachment.ContentType,
		Url:         out.Attachment.URL,
		SizeBytes:   out.Attachment.SizeBytes,
	}}, nil
}

func (s *JobServer) DeleteJobAttachment(ctx context.Context, req *jobv1.DeleteJobAttachmentRequest) (*jobv1.DeleteJobAttachmentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}

	out, err := s.DeleteAttachmentUC.Execute(ctx, application.DeleteJobAttachmentInput{
		JobID:        req.JobId,
		AttachmentID: req.AttachmentId,
		ClientID:     callerID,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &jobv1.DeleteJobAttachmentResponse{Deleted: out.Deleted}, nil
}

func (s *JobServer) InviteFreelancerToJob(ctx context.Context, req *jobv1.InviteFreelancerToJobRequest) (*jobv1.InviteFreelancerToJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.InviteFreelancerUC.Execute(ctx, application.InviteFreelancerToJobInput{
		JobID:        req.JobId,
		ClientID:     callerID,
		FreelancerID: req.FreelancerId,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.InviteFreelancerToJobResponse{Invited: out.Invited}, nil
}

func (s *JobServer) ListJobApplicants(ctx context.Context, req *jobv1.ListJobApplicantsRequest) (*jobv1.ListJobApplicantsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.ListApplicantsUC.Execute(ctx, application.ListJobApplicantsInput{
		JobID:     req.JobId,
		ClientID:  callerID,
		PageSize:  req.PageSize,
		PageToken: req.PageToken,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	applicants := make([]*jobv1.Applicant, 0, len(out.Applicants))
	for _, a := range out.Applicants {
		applicants = append(applicants, &jobv1.Applicant{
			ProposalId:    a.ProposalID,
			FreelancerId:  a.FreelancerID,
			Stage:         applicantStageToEnum(a.Stage),
			ConnectsSpent: a.ConnectsSpent,
		})
	}
	return &jobv1.ListJobApplicantsResponse{Applicants: applicants, NextPageToken: out.NextPageToken}, nil
}

func (s *JobServer) SetApplicantStage(ctx context.Context, req *jobv1.SetApplicantStageRequest) (*jobv1.SetApplicantStageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	stage, mapErr := applicantStageFromEnum(req.Stage)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}
	out, err := s.SetApplicantUC.Execute(ctx, application.SetApplicantStageInput{
		ProposalID: req.ProposalId,
		ClientID:   callerID,
		Stage:      stage,
		Reason:     req.Reason,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.SetApplicantStageResponse{Updated: out.Updated}, nil
}

func (s *JobServer) SetJobVisibility(ctx context.Context, req *jobv1.SetJobVisibilityRequest) (*jobv1.SetJobVisibilityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	visibility, mapErr := visibilityFromEnum(req.Visibility)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}
	out, err := s.SetVisibilityUC.Execute(ctx, application.SetJobVisibilityInput{JobID: req.JobId, ClientID: callerID, Visibility: visibility})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.SetJobVisibilityResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) SetJobBudgetRange(ctx context.Context, req *jobv1.SetJobBudgetRangeRequest) (*jobv1.SetJobBudgetRangeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.SetBudgetRangeUC.Execute(ctx, application.SetJobBudgetRangeInput{JobID: req.JobId, ClientID: callerID, BudgetMin: req.BudgetMin, BudgetMax: req.BudgetMax})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.SetJobBudgetRangeResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) PauseJob(ctx context.Context, req *jobv1.PauseJobRequest) (*jobv1.PauseJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.PauseJobUC.Execute(ctx, application.PauseJobInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.PauseJobResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) ReopenJob(ctx context.Context, req *jobv1.ReopenJobRequest) (*jobv1.ReopenJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.ReopenJobUC.Execute(ctx, application.ReopenJobInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.ReopenJobResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) MarkJobFilled(ctx context.Context, req *jobv1.MarkJobFilledRequest) (*jobv1.MarkJobFilledResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.MarkFilledUC.Execute(ctx, application.MarkJobFilledInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.MarkJobFilledResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) SearchJobs(ctx context.Context, req *jobv1.SearchJobsRequest) (*jobv1.SearchJobsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	jobType := ""
	if req.JobType != jobv1.JobType_JOB_TYPE_UNSPECIFIED {
		mapped, mapErr := jobTypeFromEnum(req.JobType)
		if mapErr != nil {
			return nil, status.Error(codes.InvalidArgument, mapErr.Error())
		}
		jobType = mapped
	}
	visibility, mapErr := visibilityFromEnum(req.Visibility)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}
	out, err := s.SearchJobsUC.Execute(ctx, application.SearchJobsInput{
		PageSize:   req.PageSize,
		PageToken:  req.PageToken,
		Query:      req.Query,
		Skills:     req.Skills,
		JobType:    jobType,
		Visibility: visibility,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	jobs := make([]*jobv1.Job, 0, len(out.Jobs))
	for _, j := range out.Jobs {
		jobs = append(jobs, toProtoJob(j))
	}
	return &jobv1.SearchJobsResponse{Jobs: jobs, NextPageToken: out.NextPageToken}, nil
}

func (s *JobServer) ListJobFacets(ctx context.Context, req *jobv1.ListJobFacetsRequest) (*jobv1.ListJobFacetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.ListFacetsUC.Execute(ctx, application.ListJobFacetsInput{Query: req.Query})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.ListJobFacetsResponse{
		Skills:     toProtoFacets(out.Skills),
		JobTypes:   toProtoFacets(out.JobTypes),
		Visibility: toProtoFacets(out.Visibility),
		Status:     toProtoFacets(out.Status),
	}, nil
}

func (s *JobServer) ListJobAttachments(ctx context.Context, req *jobv1.ListJobAttachmentsRequest) (*jobv1.ListJobAttachmentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.ListAttachmentsUC.Execute(ctx, application.ListJobAttachmentsInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	attachments := make([]*jobv1.JobAttachment, 0, len(out.Attachments))
	for _, a := range out.Attachments {
		attachments = append(attachments, &jobv1.JobAttachment{Id: a.ID, FileName: a.FileName, ContentType: a.ContentType, Url: a.URL, SizeBytes: a.SizeBytes})
	}
	return &jobv1.ListJobAttachmentsResponse{Attachments: attachments}, nil
}

func (s *JobServer) GetJobAttachmentDownloadUrl(ctx context.Context, req *jobv1.GetJobAttachmentDownloadUrlRequest) (*jobv1.GetJobAttachmentDownloadUrlResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.GetAttachmentURLUC.Execute(ctx, application.GetJobAttachmentDownloadURLInput{JobID: req.JobId, AttachmentID: req.AttachmentId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.GetJobAttachmentDownloadUrlResponse{Url: out.URL}, nil
}

func (s *JobServer) GetPublicJobDetail(ctx context.Context, req *jobv1.GetPublicJobDetailRequest) (*jobv1.GetPublicJobDetailResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.GetPublicJobUC.Execute(ctx, application.GetPublicJobDetailInput{JobID: req.JobId})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.GetPublicJobDetailResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) ListInvitedJobs(ctx context.Context, req *jobv1.ListInvitedJobsRequest) (*jobv1.ListInvitedJobsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	out, err := s.ListInvitedJobsUC.Execute(ctx, application.ListInvitedJobsInput{FreelancerID: callerID, PageSize: req.PageSize, PageToken: req.PageToken})
	if err != nil {
		return nil, toStatus(err)
	}
	invites := make([]*jobv1.InvitedJob, 0, len(out.InvitedJobs))
	for _, ij := range out.InvitedJobs {
		invites = append(invites, &jobv1.InvitedJob{
			Job: toProtoJob(ij.Job),
			Invite: &jobv1.JobInvite{
				JobId:                ij.Invite.JobID,
				ClientId:             ij.Invite.ClientID.String(),
				FreelancerId:         ij.Invite.FreelancerID.String(),
				InvitedAtUnixSeconds: ij.Invite.InvitedAt.Unix(),
				ResponseStatus:       inviteResponseToEnum(ij.Invite.ResponseStatus),
			},
		})
	}
	return &jobv1.ListInvitedJobsResponse{Invites: invites, NextPageToken: out.NextPageToken}, nil
}

func (s *JobServer) RespondToJobInvite(ctx context.Context, req *jobv1.RespondToJobInviteRequest) (*jobv1.RespondToJobInviteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	state, mapErr := inviteResponseFromEnum(req.ResponseStatus)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}
	out, err := s.RespondInviteUC.Execute(ctx, application.RespondToJobInviteInput{JobID: req.JobId, FreelancerID: callerID, ResponseState: state})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.RespondToJobInviteResponse{Updated: out.Updated}, nil
}

func (s *JobServer) SaveJob(ctx context.Context, req *jobv1.SaveJobRequest) (*jobv1.SaveJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	out, err := s.SaveJobUC.Execute(ctx, application.SaveJobInput{JobID: req.JobId, FreelancerID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.SaveJobResponse{Saved: out.Saved}, nil
}

func (s *JobServer) UnsaveJob(ctx context.Context, req *jobv1.UnsaveJobRequest) (*jobv1.UnsaveJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	out, err := s.UnsaveJobUC.Execute(ctx, application.UnsaveJobInput{JobID: req.JobId, FreelancerID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.UnsaveJobResponse{Removed: out.Removed}, nil
}

func (s *JobServer) ListSavedJobs(ctx context.Context, req *jobv1.ListSavedJobsRequest) (*jobv1.ListSavedJobsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	out, err := s.ListSavedJobsUC.Execute(ctx, application.ListSavedJobsInput{FreelancerID: callerID, PageSize: req.PageSize, PageToken: req.PageToken})
	if err != nil {
		return nil, toStatus(err)
	}
	jobs := make([]*jobv1.Job, 0, len(out.Jobs))
	for _, j := range out.Jobs {
		jobs = append(jobs, toProtoJob(j))
	}
	return &jobv1.ListSavedJobsResponse{Jobs: jobs, NextPageToken: out.NextPageToken}, nil
}

func (s *JobServer) HireApplicant(ctx context.Context, req *jobv1.HireApplicantRequest) (*jobv1.HireApplicantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.HireApplicantUC.Execute(ctx, application.HireApplicantInput{ProposalID: req.ProposalId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.HireApplicantResponse{Hired: out.Hired, JobId: out.JobID}, nil
}

func (s *JobServer) RejectAllApplicants(ctx context.Context, req *jobv1.RejectAllApplicantsRequest) (*jobv1.RejectAllApplicantsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.RejectAllUC.Execute(ctx, application.RejectAllApplicantsInput{JobID: req.JobId, ClientID: callerID, Reason: req.Reason})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.RejectAllApplicantsResponse{RejectedCount: out.RejectedCount}, nil
}

func (s *JobServer) ReopenHiringForJob(ctx context.Context, req *jobv1.ReopenHiringForJobRequest) (*jobv1.ReopenHiringForJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.ReopenHiringUC.Execute(ctx, application.ReopenHiringForJobInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.ReopenHiringForJobResponse{Job: toProtoJob(out.Job)}, nil
}

func (s *JobServer) GetJobStats(ctx context.Context, req *jobv1.GetJobStatsRequest) (*jobv1.GetJobStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.GetJobStatsUC.Execute(ctx, application.GetJobStatsInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.GetJobStatsResponse{
		InviteCount:         out.InviteCount,
		InviteAcceptedCount: out.InviteAcceptedCount,
		InviteDeclinedCount: out.InviteDeclinedCount,
		ApplicantCount:      out.ApplicantCount,
		ShortlistedCount:    out.ShortlistedCount,
		RejectedCount:       out.RejectedCount,
		HiredCount:          out.HiredCount,
	}, nil
}

func (s *JobServer) SearchJobsV2(ctx context.Context, req *jobv1.SearchJobsV2Request) (*jobv1.SearchJobsV2Response, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	jobType := ""
	if req.JobType != jobv1.JobType_JOB_TYPE_UNSPECIFIED {
		mapped, mapErr := jobTypeFromEnum(req.JobType)
		if mapErr != nil {
			return nil, status.Error(codes.InvalidArgument, mapErr.Error())
		}
		jobType = mapped
	}
	visibility, mapErr := visibilityFromEnum(req.Visibility)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}
	sortBy := sortByFromEnum(req.SortBy)
	out, err := s.SearchJobsV2UC.Execute(ctx, application.SearchJobsV2Input{
		PageSize:   req.PageSize,
		PageToken:  req.PageToken,
		Query:      req.Query,
		Skills:     req.Skills,
		JobType:    jobType,
		Visibility: visibility,
		SortBy:     sortBy,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	jobs := make([]*jobv1.Job, 0, len(out.Jobs))
	for _, j := range out.Jobs {
		jobs = append(jobs, toProtoJob(j))
	}
	return &jobv1.SearchJobsV2Response{Jobs: jobs, NextPageToken: out.NextPageToken}, nil
}

func (s *JobServer) MarkJobCompleted(ctx context.Context, req *jobv1.MarkJobCompletedRequest) (*jobv1.MarkJobCompletedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.MarkCompletedUC.Execute(ctx, application.MarkJobCompletedInput{JobID: req.JobId, ClientID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.MarkJobCompletedResponse{Completed: out.Completed}, nil
}

func (s *JobServer) CancelJobWithSettlementPolicy(ctx context.Context, req *jobv1.CancelJobWithSettlementPolicyRequest) (*jobv1.CancelJobWithSettlementPolicyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	policy, mapErr := settlementPolicyFromEnum(req.SettlementPolicy)
	if mapErr != nil {
		return nil, status.Error(codes.InvalidArgument, mapErr.Error())
	}
	out, err := s.CancelWithSettleUC.Execute(ctx, application.CancelJobWithSettlementPolicyInput{JobID: req.JobId, ClientID: callerID, SettlementPolicy: policy, Reason: req.Reason})
	if err != nil {
		return nil, toStatus(err)
	}
	return &jobv1.CancelJobWithSettlementPolicyResponse{Canceled: out.Canceled}, nil
}

func toProtoJob(in domain.Job) *jobv1.Job {
	attachments := make([]*jobv1.JobAttachment, 0, len(in.Attachments))
	for _, a := range in.Attachments {
		attachments = append(attachments, &jobv1.JobAttachment{
			Id:          a.ID,
			FileName:    a.FileName,
			ContentType: a.ContentType,
			Url:         a.URL,
			SizeBytes:   a.SizeBytes,
		})
	}

	out := &jobv1.Job{
		Id:                   in.ID,
		ClientId:             in.ClientID.String(),
		Title:                in.Title,
		Description:          in.Description,
		RequiredSkills:       in.RequiredSkills,
		BudgetFixed:          in.BudgetFixed,
		HourlyRate:           in.HourlyRate,
		Currency:             in.Currency,
		BudgetMin:            in.BudgetMin,
		BudgetMax:            in.BudgetMax,
		Attachments:          attachments,
		JobTypeEnum:          jobTypeToEnum(in.JobType),
		StatusEnum:           jobStatusToEnum(in.Status),
		Visibility:           visibilityToEnum(in.Visibility),
		CloseReason:          closeReasonToEnum(in.CloseReason),
		SettlementPolicy:     settlementPolicyToEnum(in.SettlementPolicy),
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: in.UpdatedAt.Unix(),
	}
	if in.Deadline != nil {
		out.DeadlineUnixSeconds = in.Deadline.Unix()
	}
	if in.ClosedAt != nil {
		out.ClosedAtUnixSeconds = in.ClosedAt.Unix()
	}
	if in.PausedAt != nil {
		out.PausedAtUnixSeconds = in.PausedAt.Unix()
	}
	if in.FilledAt != nil {
		out.FilledAtUnixSeconds = in.FilledAt.Unix()
	}
	if in.CompletedAt != nil {
		out.CompletedAtUnixSeconds = in.CompletedAt.Unix()
	}
	if in.CanceledAt != nil {
		out.CanceledAtUnixSeconds = in.CanceledAt.Unix()
	}
	return out
}

func jobTypeFromEnum(in jobv1.JobType) (string, error) {
	switch in {
	case jobv1.JobType_JOB_TYPE_FIXED:
		return domain.JobTypeFixed, nil
	case jobv1.JobType_JOB_TYPE_HOURLY:
		return domain.JobTypeHourly, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid job_type_enum")
	}
}

func jobTypeToEnum(in string) jobv1.JobType {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case domain.JobTypeFixed:
		return jobv1.JobType_JOB_TYPE_FIXED
	case domain.JobTypeHourly:
		return jobv1.JobType_JOB_TYPE_HOURLY
	default:
		return jobv1.JobType_JOB_TYPE_UNSPECIFIED
	}
}

func jobStatusFromEnum(in jobv1.JobStatus) string {
	switch in {
	case jobv1.JobStatus_JOB_STATUS_OPEN:
		return domain.JobStatusOpen
	case jobv1.JobStatus_JOB_STATUS_CLOSED:
		return domain.JobStatusClosed
	case jobv1.JobStatus_JOB_STATUS_PAUSED:
		return domain.JobStatusPaused
	case jobv1.JobStatus_JOB_STATUS_FILLED:
		return domain.JobStatusFilled
	case jobv1.JobStatus_JOB_STATUS_COMPLETED:
		return domain.JobStatusCompleted
	case jobv1.JobStatus_JOB_STATUS_CANCELED:
		return domain.JobStatusCanceled
	default:
		return ""
	}
}

func jobStatusToEnum(in string) jobv1.JobStatus {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case domain.JobStatusOpen:
		return jobv1.JobStatus_JOB_STATUS_OPEN
	case domain.JobStatusClosed:
		return jobv1.JobStatus_JOB_STATUS_CLOSED
	case domain.JobStatusPaused:
		return jobv1.JobStatus_JOB_STATUS_PAUSED
	case domain.JobStatusFilled:
		return jobv1.JobStatus_JOB_STATUS_FILLED
	case domain.JobStatusCompleted:
		return jobv1.JobStatus_JOB_STATUS_COMPLETED
	case domain.JobStatusCanceled:
		return jobv1.JobStatus_JOB_STATUS_CANCELED
	default:
		return jobv1.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}

func visibilityFromEnum(in jobv1.Visibility) (string, error) {
	switch in {
	case jobv1.Visibility_VISIBILITY_UNSPECIFIED:
		return "", nil
	case jobv1.Visibility_VISIBILITY_PUBLIC:
		return domain.VisibilityPublic, nil
	case jobv1.Visibility_VISIBILITY_PRIVATE:
		return domain.VisibilityPrivate, nil
	case jobv1.Visibility_VISIBILITY_INVITE_ONLY:
		return domain.VisibilityInviteOnly, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid visibility")
	}
}

func visibilityToEnum(in string) jobv1.Visibility {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case domain.VisibilityPublic:
		return jobv1.Visibility_VISIBILITY_PUBLIC
	case domain.VisibilityPrivate:
		return jobv1.Visibility_VISIBILITY_PRIVATE
	case domain.VisibilityInviteOnly:
		return jobv1.Visibility_VISIBILITY_INVITE_ONLY
	default:
		return jobv1.Visibility_VISIBILITY_UNSPECIFIED
	}
}

func closeReasonFromEnum(in jobv1.CloseReason) (string, error) {
	switch in {
	case jobv1.CloseReason_CLOSE_REASON_UNSPECIFIED:
		return "", nil
	case jobv1.CloseReason_CLOSE_REASON_CANCELED:
		return domain.CloseReasonCanceled, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid reason_enum")
	}
}

func closeReasonToEnum(in string) jobv1.CloseReason {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case domain.CloseReasonCanceled:
		return jobv1.CloseReason_CLOSE_REASON_CANCELED
	default:
		return jobv1.CloseReason_CLOSE_REASON_UNSPECIFIED
	}
}

func settlementPolicyToEnum(in string) jobv1.SettlementPolicy {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case application.SettlementPolicyRefundRemaining:
		return jobv1.SettlementPolicy_SETTLEMENT_POLICY_REFUND_REMAINING
	case application.SettlementPolicyNoRefund:
		return jobv1.SettlementPolicy_SETTLEMENT_POLICY_NO_REFUND
	default:
		return jobv1.SettlementPolicy_SETTLEMENT_POLICY_UNSPECIFIED
	}
}

func applicantStageFromEnum(in jobv1.ApplicantStage) (string, error) {
	switch in {
	case jobv1.ApplicantStage_APPLICANT_STAGE_SENT:
		return application.ApplicantStageSent, nil
	case jobv1.ApplicantStage_APPLICANT_STAGE_SHORTLISTED:
		return application.ApplicantStageShortlisted, nil
	case jobv1.ApplicantStage_APPLICANT_STAGE_REJECTED:
		return application.ApplicantStageRejected, nil
	case jobv1.ApplicantStage_APPLICANT_STAGE_HIRED:
		return application.ApplicantStageHired, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid applicant stage")
	}
}

func applicantStageToEnum(in string) jobv1.ApplicantStage {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case application.ApplicantStageShortlisted:
		return jobv1.ApplicantStage_APPLICANT_STAGE_SHORTLISTED
	case application.ApplicantStageRejected:
		return jobv1.ApplicantStage_APPLICANT_STAGE_REJECTED
	case application.ApplicantStageHired:
		return jobv1.ApplicantStage_APPLICANT_STAGE_HIRED
	default:
		return jobv1.ApplicantStage_APPLICANT_STAGE_SENT
	}
}

func inviteResponseFromEnum(in jobv1.InviteResponseStatus) (string, error) {
	switch in {
	case jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_ACCEPTED:
		return application.InviteResponseAccepted, nil
	case jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_DECLINED:
		return application.InviteResponseDeclined, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid invite response status")
	}
}

func inviteResponseToEnum(in string) jobv1.InviteResponseStatus {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case application.InviteResponseAccepted:
		return jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_ACCEPTED
	case application.InviteResponseDeclined:
		return jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_DECLINED
	default:
		return jobv1.InviteResponseStatus_INVITE_RESPONSE_STATUS_UNSPECIFIED
	}
}

func sortByFromEnum(in jobv1.JobSortBy) string {
	switch in {
	case jobv1.JobSortBy_JOB_SORT_BY_NEWEST:
		return "newest"
	case jobv1.JobSortBy_JOB_SORT_BY_OLDEST:
		return "oldest"
	case jobv1.JobSortBy_JOB_SORT_BY_BUDGET_HIGH:
		return "budget_high"
	case jobv1.JobSortBy_JOB_SORT_BY_BUDGET_LOW:
		return "budget_low"
	default:
		return "relevance"
	}
}

func settlementPolicyFromEnum(in jobv1.SettlementPolicy) (string, error) {
	switch in {
	case jobv1.SettlementPolicy_SETTLEMENT_POLICY_REFUND_REMAINING:
		return application.SettlementPolicyRefundRemaining, nil
	case jobv1.SettlementPolicy_SETTLEMENT_POLICY_NO_REFUND:
		return application.SettlementPolicyNoRefund, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid settlement policy")
	}
}

func toProtoFacets(in []application.FacetBucket) []*jobv1.FacetValue {
	out := make([]*jobv1.FacetValue, 0, len(in))
	for _, f := range in {
		out = append(out, &jobv1.FacetValue{Value: f.Value, Count: f.Count})
	}
	return out
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, err.Error())
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"), strings.Contains(msg, "too long"), strings.Contains(msg, "must"):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
