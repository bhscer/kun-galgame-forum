package dto

import (
	"kun-galgame-api/internal/toolset/model"
	userModel "kun-galgame-api/internal/user/model"
)

// ──────────────────────────────────────────
// Requests
// ──────────────────────────────────────────

// Wire field name is `toolsetResourceId` to match the frontend convention
// (validations/toolset.ts) and the legacy nitro endpoints. The internal Go
// field stays `ResourceID` since the service only deals with one kind of
// resource id at a time.

type ResourceDetailRequest struct {
	ResourceID int `query:"toolsetResourceId" validate:"required,min=1"`
}

type CreateResourceRequest struct {
	Content  string `json:"content" validate:"max=1007"`
	Type     string `json:"type" validate:"required,oneof=s3 user"`
	Code     string `json:"code" validate:"max=1007"`
	Password string `json:"password" validate:"max=1007"`
	Size     string `json:"size" validate:"max=107"`
	Note     string `json:"note" validate:"max=1007"`
}

type UpdateResourceRequest struct {
	ResourceID int    `json:"toolsetResourceId" validate:"required,min=1"`
	Content    string `json:"content" validate:"max=1007"`
	Code       string `json:"code" validate:"max=1007"`
	Password   string `json:"password" validate:"max=1007"`
	Size       string `json:"size" validate:"max=107"`
	Note       string `json:"note" validate:"max=1007"`
}

type DeleteResourceRequest struct {
	ResourceID int `query:"toolsetResourceId" validate:"required,min=1"`
}

// ──────────────────────────────────────────
// Responses
// ──────────────────────────────────────────

// ResourceDetailResponse is returned by GET /toolset/:id/resource/detail.
// Wire format is flat — model fields appear at the JSON top level via
// struct embedding, with `user` joined as a sibling key. This matches
// the pre-refactor nitro response (which used `prisma.include` to merge
// the relation) and the frontend's ToolsetResourceDetail interface
// (which extends ToolsetResource flatly). Re-nesting the model under a
// `resource` key would silently break the frontend: content/created
// would be undefined and downstream UI prints "NaN 年前" / "/undefined".
type ResourceDetailResponse struct {
	model.GalgameToolsetResource
	User userModel.UserBrief `json:"user"`
}

// CreatedResourceResponse is the resource row returned by POST.
// (Handler returns the model directly; we preserve that.)
type CreatedResourceResponse = model.GalgameToolsetResource
