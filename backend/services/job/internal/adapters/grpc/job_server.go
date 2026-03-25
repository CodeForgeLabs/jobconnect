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
	UpdateJobUC        *application.UpdateJob
	ListMyJobsUC       *application.ListMyJobs
	ListOpenJobsUC     *application.ListOpenJobs
	CloseJobUC         *application.CloseJob
	UploadAttachmentUC *application.UploadJobAttachment
	DeleteAttachmentUC *application.DeleteJobAttachment
	TokenParser        TokenParser
}

func NewJobServer(
	createJob *application.CreateJob,
	getJob *application.GetJob,
	updateJob *application.UpdateJob,
	listMyJobs *application.ListMyJobs,
	listOpenJobs *application.ListOpenJobs,
	closeJob *application.CloseJob,
	uploadAttachment *application.UploadJobAttachment,
	deleteAttachment *application.DeleteJobAttachment,
	tokenParser TokenParser,
) *JobServer {
	return &JobServer{
		CreateJobUC:        createJob,
		GetJobUC:           getJob,
		UpdateJobUC:        updateJob,
		ListMyJobsUC:       listMyJobs,
		ListOpenJobsUC:     listOpenJobs,
		CloseJobUC:         closeJob,
		UploadAttachmentUC: uploadAttachment,
		DeleteAttachmentUC: deleteAttachment,
		TokenParser:        tokenParser,
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
	out, err := s.CreateJobUC.Execute(ctx, application.CreateJobInput{
		ClientID:       callerID,
		Title:          req.Title,
		Description:    req.Description,
		RequiredSkills: req.RequiredSkills,
		JobType:        req.JobType,
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
		JobID:    req.JobId,
		ClientID: callerID,
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
	if req.JobType != nil {
		in.JobType = req.JobType
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
		Status:    req.Status,
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
	out, err := s.ListOpenJobsUC.Execute(ctx, application.ListOpenJobsInput{
		PageSize:    req.PageSize,
		PageToken:   req.PageToken,
		SearchQuery: req.SearchQuery,
		Skills:      req.Skills,
		JobType:     req.JobType,
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
	out, err := s.CloseJobUC.Execute(ctx, application.CloseJobInput{JobID: req.JobId, ClientID: callerID, Reason: req.Reason})
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
		JobType:              in.JobType,
		BudgetFixed:          in.BudgetFixed,
		HourlyRate:           in.HourlyRate,
		Currency:             in.Currency,
		Attachments:          attachments,
		Status:               in.Status,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: in.UpdatedAt.Unix(),
	}
	if in.Deadline != nil {
		out.DeadlineUnixSeconds = in.Deadline.Unix()
	}
	if in.ClosedAt != nil {
		out.ClosedAtUnixSeconds = in.ClosedAt.Unix()
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
