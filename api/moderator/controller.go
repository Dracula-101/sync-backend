package moderator

import (
	"sync-backend/api/community"
	"sync-backend/api/moderator/dto"
	"sync-backend/api/moderator/model"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	"sync-backend/arch/middleware"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type moderatorController struct {
	network.BaseController
	common.ContextPayload
	logger                utils.AppLogger
	authenticatorProvider network.AuthenticationProvider
	uploadProvider        middleware.UploadProvider
	moderatorService      ModeratorService
	communityService      community.CommunityService
	userService           user.UserService
}

// NewModeratorController creates a new moderator controller
func NewModeratorController(
	authenticatorProvider network.AuthenticationProvider,
	uploadProvider middleware.UploadProvider,
	moderatorService ModeratorService,
	communityService community.CommunityService,
	userService user.UserService,
) network.Controller {
	return &moderatorController{
		logger:                utils.NewServiceLogger("ModeratorController"),
		BaseController:        network.NewBaseController("/moderator", authenticatorProvider),
		ContextPayload:        common.NewContextPayload(),
		authenticatorProvider: authenticatorProvider,
		uploadProvider:        uploadProvider,
		moderatorService:      moderatorService,
		communityService:      communityService,
		userService:           userService,
	}
}

// MountRoutes mounts the routes for the moderator controller
func (c *moderatorController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting moderator routes")
	group.Use(c.authenticatorProvider.Middleware())

	// Moderator management routes
	group.POST("/community/:communityId/add", c.AddModerator)
	group.DELETE("/community/:communityId/remove/:userId", c.RemoveModerator)
	group.PATCH("/community/:communityId/update/:userId", c.UpdateModerator)
	group.GET("/community/:communityId/list", c.ListModerators)
	group.GET("/community/:communityId/get/:userId", c.GetModerator)

	// Permission check routes
	group.GET("/community/:communityId/check-permission/:permission", c.CheckModeratorPermission)

	// User moderation routes
	group.POST("/community/:communityId/ban/:userId", c.BanUser)
	group.POST("/community/:communityId/unban/:userId", c.UnbanUser)

	// Report system routes
	group.POST("/report/create", c.CreateReport)
	group.PATCH("/report/:reportId/process", c.ProcessReport)
	group.GET("/report/:reportId", c.GetReport)
	group.GET("/community/:communityId/reports", c.ListReports)

	// Moderation log routes
	group.GET("/community/:communityId/logs", c.GetModLogs)
}

// AddModerator handles adding a new moderator to a community
func (c *moderatorController) AddModerator(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			nil,
		)
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	body, err := network.ReqBody(ctx, dto.NewAddModeratorRequest())
	if err != nil {
		return
	}

	// Convert string permissions to ModeratorPermission type if provided
	var permissions []model.ModeratorPermission
	if len(body.Permissions) > 0 {
		permissions = make([]model.ModeratorPermission, len(body.Permissions))
		for i, p := range body.Permissions {
			permissions[i] = model.ModeratorPermission(p)
		}
	}

	moderator, apiErr := c.moderatorService.AddModerator(body.UserId, communityId, body.Role, *userId)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	// add moderator info to community
	apiErr = c.communityService.AddModerator(communityId, body.UserId, *userId)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}
	// add moderator info to the user
	err = c.userService.AddModerator(body.UserId, communityId)
	if err != nil {
		c.Send(ctx).InternalServerError(
			"Failed to add moderator info to user",
			"An error occurred while adding moderator info to the user",
			network.DB_ERROR,
			err,
		)
		return
	}

	c.Send(ctx).SuccessDataResponse("Moderator added successfully", moderator)
}

// RemoveModerator handles removing a moderator from a community
func (c *moderatorController) RemoveModerator(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	userId := ctx.Param("userId")
	if communityId == "" || userId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID and User ID are required",
			"Please provide valid Community ID and User ID in the request params",
			nil,
		)
		return
	}

	requesterId := c.ContextPayload.MustGetUserId(ctx)
	body, err := network.ReqBody(ctx, dto.NewRemoveModeratorRequest())
	if err != nil {
		return
	}

	apiErr := c.moderatorService.RemoveModerator(communityId, userId, *requesterId, body.Reason)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	// Remove moderator info from community
	err = c.communityService.RemoveModerator(communityId, userId, *requesterId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	insertErr := c.userService.RemoveModerator(userId, communityId)
	if insertErr != nil {
		c.Send(ctx).InternalServerError(
			"Failed to remove moderator info from user",
			"An error occurred while removing moderator info from the user",
			network.DB_ERROR,
			insertErr,
		)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Moderator removed successfully")
}

// UpdateModerator handles updating a moderator's role or permissions
func (c *moderatorController) UpdateModerator(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	userId := ctx.Param("userId")
	if communityId == "" || userId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID and User ID are required",
			"Please provide valid Community ID and User ID in the request params",
			nil,
		)
		return
	}

	requesterId := c.ContextPayload.MustGetUserId(ctx)
	body, err := network.ReqBody(ctx, dto.NewUpdateModeratorRequest())
	if err != nil {
		return
	}

	// Convert string permissions to ModeratorPermission type if provided
	var permissions []model.ModeratorPermission
	if len(body.Permissions) > 0 {
		permissions = make([]model.ModeratorPermission, len(body.Permissions))
		for i, p := range body.Permissions {
			permissions[i] = model.ModeratorPermission(p)
		}
	}

	var status *string
	if body.Status != "" {
		status = &body.Status
	}

	var role *model.ModeratorRole
	if body.Role != "" {
		r := body.Role
		role = &r
	}

	moderator, apiErr := c.moderatorService.UpdateModerator(communityId, userId, *requesterId, role, permissions, status, &body.Notes)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("Moderator updated successfully", moderator)
}

// GetModerator handles getting details of a specific moderator
func (c *moderatorController) GetModerator(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	userId := ctx.Param("userId")
	if communityId == "" || userId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID and User ID are required",
			"Please provide valid Community ID and User ID in the request params",
			nil,
		)
		return
	}

	moderator, apiErr := c.moderatorService.GetModerator(communityId, userId)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("Moderator details retrieved successfully", moderator)
}

// ListModerators handles listing all moderators for a community
func (c *moderatorController) ListModerators(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			nil,
		)
		return
	}

	// Parse pagination from query
	query, err := network.ReqQuery(ctx, dto.NewListModeratorsRequest())
	if err != nil {
		return
	}

	moderators, total, apiErr := c.moderatorService.ListModerators(communityId, query.Page, query.Limit)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse(
		"Moderators retrieved successfully",
		dto.NewListModeratorsResponse(moderators, query.Page, query.Limit, total),
	)
}

// CheckModeratorPermission handles checking if a user has a specific permission
func (c *moderatorController) CheckModeratorPermission(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	permission := ctx.Param("permission")
	if communityId == "" || permission == "" {
		c.Send(ctx).BadRequestError(
			"Community ID and Permission are required",
			"Please provide valid Community ID and Permission in the request params",
			nil,
		)
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	hasPermission, apiErr := c.moderatorService.HasModeratorPermission(*userId, communityId, model.ModeratorPermission(permission))
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	response := dto.NewPermissionCheckResponse(hasPermission)
	c.Send(ctx).SuccessDataResponse("Permission check completed", response)
}

// CreateReport handles creating a report
func (c *moderatorController) CreateReport(ctx *gin.Context) {
	userId := c.ContextPayload.MustGetUserId(ctx)
	body, err := network.ReqBody(ctx, dto.NewCreateReportRequest())
	if err != nil {
		return
	}

	report, apiErr := c.moderatorService.CreateReport(
		*userId,
		body.CommunityId,
		body.TargetId,
		model.ReportType(body.TargetType),
		model.ReportReason(body.Reason),
		body.Description,
	)

	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("Report created successfully", report)
}

// ProcessReport handles processing a report
func (c *moderatorController) ProcessReport(ctx *gin.Context) {
	reportId := ctx.Param("reportId")
	if reportId == "" {
		c.Send(ctx).BadRequestError(
			"Report ID is required",
			"Please provide a valid report ID in the request params",
			nil,
		)
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	body, err := network.ReqBody(ctx, dto.NewProcessReportRequest())
	if err != nil {
		return
	}

	report, apiErr := c.moderatorService.ProcessReport(
		reportId,
		*userId,
		model.ReportStatus(body.Status),
		body.ModeratorNotes,
		body.ActionTaken,
	)

	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("Report processed successfully", report)
}

// GetReport handles getting details of a specific report
func (c *moderatorController) GetReport(ctx *gin.Context) {
	reportId := ctx.Param("reportId")
	if reportId == "" {
		c.Send(ctx).BadRequestError(
			"Report ID is required",
			"Please provide a valid report ID in the request params",
			nil,
		)
		return
	}

	report, apiErr := c.moderatorService.GetReport(reportId)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("Report details retrieved successfully", report)
}

// ListReports handles listing reports for a community
func (c *moderatorController) ListReports(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			nil,
		)
		return
	}

	query, err := network.ReqQuery(ctx, dto.NewListReportsRequest())
	if err != nil {
		return
	}

	// Convert string values to appropriate model types
	var status *model.ReportStatus
	if query.Status != "" {
		s := model.ReportStatus(query.Status)
		status = &s
	}

	var targetType *model.ReportType
	if query.TargetType != "" {
		t := model.ReportType(query.TargetType)
		targetType = &t
	}

	reports, total, apiErr := c.moderatorService.ListReports(
		communityId,
		status,
		targetType,
		query.Page,
		query.Limit,
	)

	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse(
		"Reports retrieved successfully",
		dto.NewListReportsResponse(reports, query.Page, query.Limit, total),
	)
}

// GetModLogs handles getting moderation logs for a community
func (c *moderatorController) GetModLogs(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			nil,
		)
		return
	}

	query, err := network.ReqQuery(ctx, dto.NewListModLogsRequest())
	if err != nil {
		return
	}

	moderatorId := query.ModeratorId // Can be empty for all moderators

	logs, total, apiErr := c.moderatorService.GetModLogs(communityId, moderatorId, query.Page, query.Limit)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse(
		"Moderation logs retrieved successfully",
		dto.NewListModLogsResponse(logs, query.Page, query.Limit, total),
	)
}

// BanUser handles banning a user from a community
func (c *moderatorController) BanUser(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	userId := ctx.Param("userId")
	if communityId == "" || userId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID and User ID are required",
			"Please provide valid Community ID and User ID in the request params",
			nil,
		)
		return
	}

	moderatorId := c.ContextPayload.MustGetUserId(ctx)
	body, err := network.ReqBody(ctx, dto.NewBanUserRequest())
	if err != nil {
		return
	}

	// Convert duration
	var duration *int
	if body.Duration > 0 {
		d := int(body.Duration)
		duration = &d
	}

	modLog, apiErr := c.moderatorService.BanUser(*moderatorId, userId, communityId, body.Reason, duration)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("User banned successfully", modLog)
}

// UnbanUser handles removing a ban on a user
func (c *moderatorController) UnbanUser(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	userId := ctx.Param("userId")
	if communityId == "" || userId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID and User ID are required",
			"Please provide valid Community ID and User ID in the request params",
			nil,
		)
		return
	}

	moderatorId := c.ContextPayload.MustGetUserId(ctx)

	modLog, apiErr := c.moderatorService.UnbanUser(*moderatorId, userId, communityId)
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("User unbanned successfully", modLog)
}
