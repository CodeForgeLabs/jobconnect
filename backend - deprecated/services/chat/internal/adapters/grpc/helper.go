package grpcadapter

import (
	chatv1 "jobconnect/chat/gen/chat/v1"
	"jobconnect/chat/internal/domain"

	"time"
)

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func getMessageContentFromProto(pbContent *chatv1.MessageContent) domain.MessageContent {
	if pbContent == nil {
		return domain.MessageContent{}
	}

	content := domain.MessageContent{}

	switch t := pbContent.Type.(type) {
	case *chatv1.MessageContent_Text:
		content.Type = domain.TypeText
		content.Text = t.Text.GetText()
	case *chatv1.MessageContent_Image:
		content.Type = domain.TypeImage
		content.ImageUrl = t.Image.GetImageUrl()
		content.Caption = t.Image.GetCaption()
	case *chatv1.MessageContent_Video:
		content.Type = domain.TypeVideo
		content.VideoUrl = t.Video.GetVideoUrl()
		content.Caption = t.Video.GetCaption()
	}

	return content
}

func mapDomainToProtoContent(domainContent domain.MessageContent) *chatv1.MessageContent {
	pbContent := &chatv1.MessageContent{}

	switch domainContent.Type {
	case domain.TypeText:
		pbContent.Type = &chatv1.MessageContent_Text{
			Text: &chatv1.TextContent{Text: domainContent.Text},
		}
	case domain.TypeImage:
		pbContent.Type = &chatv1.MessageContent_Image{
			Image: &chatv1.ImageContent{ImageUrl: domainContent.ImageUrl, Caption: domainContent.Caption},
		}
	case domain.TypeVideo:
		pbContent.Type = &chatv1.MessageContent_Video{
			Video: &chatv1.VideoContent{VideoUrl: domainContent.VideoUrl, Caption: domainContent.Caption},
		}
	}

	return pbContent
}
