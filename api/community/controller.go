package community

import (
	"fmt"
	"strings"
	communitydto "sync-backend/api/community/dto/community_action"
	moderatordto "sync-backend/api/community/dto/moderation_action"
	reportdto "sync-backend/api/community/dto/report_action"
	"sync-backend/api/community/model"
	"sync-backend/api/moderator"
	moderatorModel "sync-backend/api/moderator/model"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type CommunityController interface {
	CreateCommunity(request *communitydto.CreateCommunityRequest) (*communitydto.CreateCommunityResponse, error)
	GetCommunityById(id string) (*model.Community, error)
	SearchCommunities(query string) ([]model.Community, error)
	GetMyCommunities(userId string) ([]model.Community, error)
}

type communityController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authProvider     network.AuthenticationProvider
	uploadProvider   coreMW.UploadProvider
	userService      user.UserService
	communityService CommunityService
	moderatorService moderator.ModeratorService
}

func NewCommunityController(
	authProvider network.AuthenticationProvider,
	uploadProvider coreMW.UploadProvider,
	userService user.UserService,
	communityService CommunityService,
	moderatorService moderator.ModeratorService,
) network.Controller {
	return &communityController{
		logger:           utils.NewServiceLogger("CommunityController"),
		BaseController:   network.NewBaseController("/community", authProvider),
		ContextPayload:   common.NewContextPayload(),
		authProvider:     authProvider,
		uploadProvider:   uploadProvider,
		communityService: communityService,
		moderatorService: moderatorService,
	}
}

func (c *communityController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting community routes")

	group.Use(c.authProvider.Middleware())

	/*	COMMUNITY ROUTES */
	group.POST("/create", c.uploadProvider.Middleware("avatar_photo", "background_photo"), c.CreateCommunity)
	group.GET("/:communityId", c.GetCommunityById)
	group.PUT("/:communityId", c.uploadProvider.Middleware("avatar_photo", "background_photo"), c.UpdateCommunity)
	group.DELETE("/:communityId", c.DeleteCommunity)

	/* Search and Trending Routes */
	group.GET("/search", c.SearchCommunities)
	group.GET("/autocomplete", c.AutocompeleteCommunities)
	group.GET("/trending", c.GetTrendingCommunities)

	/* USER COMMUNITY ROUTES */
	userGroup := group.Group("/user")
	userGroup.POST("/join/:communityId", c.JoinCommunity)
	userGroup.POST("/leave/:communityId", c.LeaveCommunity)
	userGroup.GET("/owner", c.GetMyCommunities)
	userGroup.GET("/joined", c.GetJoinedCommunities)

	/* MODERATOR ROUTES */
	moderatorGroup := group.Group("/moderator")
	moderatorGroup.POST("/:communityId/add", c.AddModerator)
	moderatorGroup.DELETE("/:communityId/remove/:userId", c.RemoveModerator)
	moderatorGroup.PATCH("/:communityId/update/:userId", c.UpdateModerator)
	moderatorGroup.GET("/:communityId/list", c.ListModerators)
	moderatorGroup.GET("/:communityId/get/:userId", c.GetModerator)

	/* MODERATOR PERMISSION CHECKS */
	moderatorGroup.GET("/:communityId/check-permission/:permission", c.CheckModeratorPermission)

	/* MODERATOR REPORTS AND LOGS */
	moderatorGroup.POST("/:communityId/ban/:userId", c.BanUser)
	moderatorGroup.POST("/:communityId/unban/:userId", c.UnbanUser)

	/* MODERATOR REPORTS */
	moderatorGroup.POST("/report/create", c.CreateReport)
	moderatorGroup.PATCH("/report/:reportId/process", c.ProcessReport)
	moderatorGroup.GET("/report/:reportId", c.GetReport)
	moderatorGroup.GET("/:communityId/reports", c.ListReports)
	moderatorGroup.GET("/:communityId/logs", c.GetModLogs)
}

func (c *communityController) CreateCommunity(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, communitydto.NewCreateCommunityRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	avatarPhoto := c.uploadProvider.GetUploadedFiles(ctx, "avatar_photo")
	backgroundPhoto := c.uploadProvider.GetUploadedFiles(ctx, "background_photo")
	if len(avatarPhoto.Files) > 0 {
		body.AvatarFilePath = avatarPhoto.Files[0].Path
	}
	if len(backgroundPhoto.Files) > 0 {
		body.BackgroundFilePath = backgroundPhoto.Files[0].Path
	}

	// Process the tagIds - already validated through custom validator
	tagIds := strings.Split(body.TagIds, ",")
	for i := range tagIds {
		tagIds[i] = strings.TrimSpace(tagIds[i])
	}

	community, err := c.communityService.CreateCommunity(body.Name, body.Description, tagIds, body.AvatarFilePath, body.BackgroundFilePath, *userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Community created successfully", communitydto.CreateCommunityResponse{
		CommunityId: community.CommunityId,
		Name:        community.Name,
		Slug:        community.Slug,
	})
	c.uploadProvider.DeleteUploadedFiles(ctx, "avatar_photo")
	c.uploadProvider.DeleteUploadedFiles(ctx, "background_photo")
}

func (c *communityController) UpdateCommunity(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, communitydto.NewUpdateCommunityRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	avatarPhoto := c.uploadProvider.GetUploadedFiles(ctx, "avatar_photo")
	backgroundPhoto := c.uploadProvider.GetUploadedFiles(ctx, "background_photo")
	if len(avatarPhoto.Files) > 0 {
		body.AvatarFilePath = avatarPhoto.Files[0].Path
	}
	if len(backgroundPhoto.Files) > 0 {
		body.BackgroundFilePath = backgroundPhoto.Files[0].Path
	}

	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			err,
		)
		return
	}

	_, err = c.communityService.UpdateCommunity(
		communityId,
		body.CommunityDescription,
		body.AvatarFilePath,
		body.BackgroundFilePath,
		*userId,
	)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("Community updated successfully")
	c.uploadProvider.DeleteUploadedFiles(ctx, "avatar_photo")
	c.uploadProvider.DeleteUploadedFiles(ctx, "background_photo")
}

func (c *communityController) DeleteCommunity(ctx *gin.Context) {
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
	err := c.communityService.DeleteCommunity(communityId, *userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("Community deleted successfully")
	c.uploadProvider.DeleteUploadedFiles(ctx, "avatar_photo")
	c.uploadProvider.DeleteUploadedFiles(ctx, "background_photo")
}

func (c *communityController) GetCommunityById(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, communitydto.NewGetCommunityRequest())
	if err != nil {
		return
	}

	community, err := c.communityService.GetCommunityById(params.Id)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Community fetched successfully", community)
}

func (c *communityController) SearchCommunities(ctx *gin.Context) {
	query, err := network.ReqQuery(ctx, communitydto.NewSearchCommunityRequest())
	if err != nil {
		return
	}

	communities, err := c.communityService.SearchCommunities(
		query.Query,
		query.Page,
		query.Limit,
		query.ShowPrivate,
	)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Communities fetched successfully", communities)
}

func (c *communityController) AutocompeleteCommunities(ctx *gin.Context) {
	query, err := network.ReqQuery(ctx, communitydto.NewAutocompleteCommunityRequest())
	if err != nil {
		return
	}

	communities, err := c.communityService.AutocompleteCommunities(query.Query, query.Page, query.Limit, query.ShowPrivate)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Communities fetched successfully", communities)
}

func (c *communityController) GetTrendingCommunities(ctx *gin.Context) {
	query, err := network.ReqQuery(ctx, communitydto.NewGetTrendingCommunitiesRequest())
	if err != nil {
		return
	}

	communities, err := c.communityService.GetTrendingCommunities(query.Page, query.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Communities fetched successfully", communities)
}

func (c *communityController) JoinCommunity(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, communitydto.NewJoinCommunityRequest())

	if err != nil {
		return
	}

	communityId := params.CommunityId
	userId := c.ContextPayload.MustGetUserId(ctx)
	user, err := c.userService.FindUserById(*userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	err = c.communityService.JoinCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	err = c.userService.JoinCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Joined community successfully")
}

func (c *communityController) LeaveCommunity(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, communitydto.NewLeaveCommunityRequest())

	if err != nil {
		return
	}

	communityId := params.CommunityId
	userId := c.ContextPayload.MustGetUserId(ctx)
	user, err := c.userService.FindUserById(*userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	err = c.communityService.LeaveCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	err = c.userService.LeaveCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Left community successfully")
}

func (c *communityController) GetMyCommunities(ctx *gin.Context) {
	body, err := network.ReqQuery(ctx, communitydto.NewGetMyCommunitiesRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	communities, err := c.communityService.GetCommunities(*userId, body.Page, body.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.logger.Debug("Communities: %+v", communities)

	if communities == nil {
		c.Send(ctx).NotFoundError(
			"No communities found",
			"No communities found for the user",
			nil,
		)
		return
	}

	var finalCommunities []model.Community
	for _, community := range communities {
		if community != nil {
			finalCommunities = append(finalCommunities, *community)
		}
	}

	c.Send(ctx).SuccessDataResponse(
		"Communities fetched successfully",
		communitydto.NewGetMyCommunitiesResponse(
			finalCommunities,
			len(communities),
		),
	)

}

func (c *communityController) GetJoinedCommunities(ctx *gin.Context) {
	body, err := network.ReqQuery(ctx, communitydto.NewJoinedCommunitiesRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	communities, err := c.communityService.GetCommunities(*userId, body.Page, body.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if communities == nil {
		c.Send(ctx).NotFoundError(
			"No communities found",
			"No communities found for the user",
			nil,
		)
		return
	}

	var finalCommunities []model.Community
	for _, community := range communities {
		if community != nil {
			finalCommunities = append(finalCommunities, *community)
		}
	}

	c.Send(ctx).SuccessDataResponse(
		"Communities fetched successfully",
		communitydto.NewJoinedCommunitiesResponse(
			finalCommunities,
			len(communities),
		),
	)
}

// AddModerator handles adding a new moderator to a community
func (c *communityController) AddModerator(ctx *gin.Context) {
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
	body, err := network.ReqBody(ctx, moderatordto.NewAddModeratorRequest())
	if err != nil {
		return
	}

	// Convert string permissions to ModeratorPermission type if provided
	var permissions []moderatorModel.ModeratorPermission
	if len(body.Permissions) > 0 {
		permissions = make([]moderatorModel.ModeratorPermission, len(body.Permissions))
		for i, p := range body.Permissions {
			permissions[i] = moderatorModel.ModeratorPermission(p)
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
func (c *communityController) RemoveModerator(ctx *gin.Context) {
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
	body, err := network.ReqBody(ctx, moderatordto.NewRemoveModeratorRequest())
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
func (c *communityController) UpdateModerator(ctx *gin.Context) {
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
	body, err := network.ReqBody(ctx, moderatordto.NewUpdateModeratorRequest())
	if err != nil {
		return
	}

	// Convert string permissions to ModeratorPermission type if provided
	var permissions []moderatorModel.ModeratorPermission
	if len(body.Permissions) > 0 {
		permissions = make([]moderatorModel.ModeratorPermission, len(body.Permissions))
		for i, p := range body.Permissions {
			permissions[i] = moderatorModel.ModeratorPermission(p)
		}
	}

	var status *string
	if body.Status != "" {
		status = &body.Status
	}

	var role *moderatorModel.ModeratorRole
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
func (c *communityController) GetModerator(ctx *gin.Context) {
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
func (c *communityController) ListModerators(ctx *gin.Context) {
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
	query, err := network.ReqQuery(ctx, moderatordto.NewListModeratorsRequest())
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
		moderatordto.NewListModeratorsResponse(moderators, query.Page, query.Limit, total),
	)
}

// CheckModeratorPermission handles checking if a user has a specific permission
func (c *communityController) CheckModeratorPermission(ctx *gin.Context) {
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
	hasPermission, apiErr := c.moderatorService.HasModeratorPermission(*userId, communityId, moderatorModel.ModeratorPermission(permission))
	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	response := moderatordto.NewPermissionCheckResponse(hasPermission)
	c.Send(ctx).SuccessDataResponse("Permission check completed", response)
}

// CreateReport handles creating a report
func (c *communityController) CreateReport(ctx *gin.Context) {
	userId := c.ContextPayload.MustGetUserId(ctx)
	body, err := network.ReqBody(ctx, reportdto.NewCreateReportRequest())
	if err != nil {
		return
	}

	report, apiErr := c.moderatorService.CreateReport(
		*userId,
		body.CommunityId,
		body.TargetId,
		moderatorModel.ReportType(body.TargetType),
		moderatorModel.ReportReason(body.Reason),
		body.Description,
	)

	if apiErr != nil {
		c.Send(ctx).MixedError(apiErr)
		return
	}

	c.Send(ctx).SuccessDataResponse("Report created successfully", report)
}

// ProcessReport handles processing a report
func (c *communityController) ProcessReport(ctx *gin.Context) {
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
	body, err := network.ReqBody(ctx, reportdto.NewProcessReportRequest())
	if err != nil {
		return
	}

	report, apiErr := c.moderatorService.ProcessReport(
		reportId,
		*userId,
		moderatorModel.ReportStatus(body.Status),
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
func (c *communityController) GetReport(ctx *gin.Context) {
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
func (c *communityController) ListReports(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			nil,
		)
		return
	}

	query, err := network.ReqQuery(ctx, reportdto.NewListReportsRequest())
	if err != nil {
		return
	}

	// Convert string values to appropriate model types
	var status *moderatorModel.ReportStatus
	if query.Status != "" {
		s := moderatorModel.ReportStatus(query.Status)
		status = &s
	}

	var targetType *moderatorModel.ReportType
	if query.TargetType != "" {
		t := moderatorModel.ReportType(query.TargetType)
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
		reportdto.NewListReportsResponse(reports, query.Page, query.Limit, total),
	)
}

// GetModLogs handles getting moderation logs for a community
func (c *communityController) GetModLogs(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			nil,
		)
		return
	}

	query, err := network.ReqQuery(ctx, moderatordto.NewListModLogsRequest())
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
		moderatordto.NewListModLogsResponse(logs, query.Page, query.Limit, total),
	)
}

// BanUser handles banning a user from a community
func (c *communityController) BanUser(ctx *gin.Context) {
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
	body, err := network.ReqBody(ctx, moderatordto.NewBanUserRequest())
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
func (c *communityController) UnbanUser(ctx *gin.Context) {
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
