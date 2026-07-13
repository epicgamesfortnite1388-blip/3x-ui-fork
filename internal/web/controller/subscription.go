package controller

import (
	"errors"
	"strconv"

	"github.com/mhsanaei/3x-ui/v3/internal/provider"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"

	"github.com/gin-gonic/gin"
)

type SubscriptionController struct {
	BaseController
	subscriptionService *service.SubscriptionService
	providerService     service.ProviderService
}

func NewSubscriptionController(g *gin.RouterGroup, registry *provider.Registry) *SubscriptionController {
	c := &SubscriptionController{
		subscriptionService: service.NewSubscriptionService(registry),
	}
	c.initRouter(g)
	return c
}

func (c *SubscriptionController) initRouter(g *gin.RouterGroup) {
	g.POST("/create", c.create)
	g.GET("/get/:id", c.get)
	g.POST("/renew/:id", c.renew)
	g.POST("/suspend/:id", c.suspend)
	g.POST("/resume/:id", c.resume)
	g.POST("/delete/:id", c.del)
	g.GET("/url/:id", c.getURL)
	g.GET("/status/:id", c.getStatus)
	g.POST("/sync/:id", c.syncUsage)
	g.GET("/list", c.list)
}

type createSubscriptionRequest struct {
	Provider          string `json:"provider"`
	Email             string `json:"email"`
	PlanName          string `json:"planName"`
	TrafficLimitBytes int64  `json:"trafficLimitBytes"`
	DeviceLimit       int    `json:"deviceLimit"`
	DurationDays      int    `json:"durationDays"`
	Notes             string `json:"notes,omitempty"`
}

func (c *SubscriptionController) create(ctx *gin.Context) {
	var req createSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonMsg(ctx, "invalid request body", err)
		return
	}
	if req.Provider == "" {
		req.Provider = "xray"
	}
	if req.Email == "" {
		jsonMsg(ctx, "email is required", errors.New("email is empty"))
		return
	}

	sub, err := c.subscriptionService.Create(ctx.Request.Context(), req.Provider, provider.CreateSubscriptionRequest{
		Email:             req.Email,
		PlanName:          req.PlanName,
		TrafficLimitBytes: req.TrafficLimitBytes,
		DeviceLimit:       req.DeviceLimit,
		DurationDays:      req.DurationDays,
		Notes:             req.Notes,
	})
	if err != nil {
		jsonMsg(ctx, "failed to create subscription", err)
		return
	}
	jsonObj(ctx, sub, nil)
}

type subscriptionQuery struct {
	Provider string `form:"provider"`
	ID       string `uri:"id"`
}

func (c *SubscriptionController) get(ctx *gin.Context) {
	id := ctx.Param("id")
	providerName := ctx.Query("provider")
	if providerName == "" {
		providerName = "xray"
	}
	sub, err := c.subscriptionService.Get(ctx.Request.Context(), providerName, id)
	if err != nil {
		jsonMsg(ctx, "subscription not found", err)
		return
	}
	jsonObj(ctx, sub, nil)
}

type renewRequest struct {
	Provider          string `json:"provider"`
	TrafficLimitBytes int64  `json:"trafficLimitBytes,omitempty"`
	DurationDays      int    `json:"durationDays"`
}

func (c *SubscriptionController) renew(ctx *gin.Context) {
	id := ctx.Param("id")
	var req renewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonMsg(ctx, "invalid request body", err)
		return
	}
	if req.Provider == "" {
		req.Provider = "xray"
	}
	err := c.subscriptionService.Renew(ctx.Request.Context(), req.Provider, id, provider.RenewRequest{
		TrafficLimitBytes: req.TrafficLimitBytes,
		DurationDays:      req.DurationDays,
	})
	if err != nil {
		jsonMsg(ctx, "failed to renew subscription", err)
		return
	}
	jsonMsg(ctx, "subscription renewed", nil)
}

type providerRequest struct {
	Provider string `json:"provider"`
}

func (c *SubscriptionController) suspend(ctx *gin.Context) {
	id := ctx.Param("id")
	var req providerRequest
	_ = ctx.ShouldBindJSON(&req)
	if req.Provider == "" {
		req.Provider = "xray"
	}
	err := c.subscriptionService.Suspend(ctx.Request.Context(), req.Provider, id)
	if err != nil {
		jsonMsg(ctx, "failed to suspend subscription", err)
		return
	}
	jsonMsg(ctx, "subscription suspended", nil)
}

func (c *SubscriptionController) resume(ctx *gin.Context) {
	id := ctx.Param("id")
	var req providerRequest
	_ = ctx.ShouldBindJSON(&req)
	if req.Provider == "" {
		req.Provider = "xray"
	}
	err := c.subscriptionService.Resume(ctx.Request.Context(), req.Provider, id)
	if err != nil {
		jsonMsg(ctx, "failed to resume subscription", err)
		return
	}
	jsonMsg(ctx, "subscription resumed", nil)
}

func (c *SubscriptionController) del(ctx *gin.Context) {
	id := ctx.Param("id")
	var req providerRequest
	_ = ctx.ShouldBindJSON(&req)
	if req.Provider == "" {
		req.Provider = "xray"
	}
	err := c.subscriptionService.Delete(ctx.Request.Context(), req.Provider, id)
	if err != nil {
		jsonMsg(ctx, "failed to delete subscription", err)
		return
	}
	jsonMsg(ctx, "subscription deleted", nil)
}

func (c *SubscriptionController) getURL(ctx *gin.Context) {
	id := ctx.Param("id")
	providerName := ctx.Query("provider")
	if providerName == "" {
		providerName = "xray"
	}
	url, err := c.subscriptionService.GetURL(ctx.Request.Context(), providerName, id)
	if err != nil {
		jsonMsg(ctx, "failed to get subscription URL", err)
		return
	}
	jsonObj(ctx, map[string]string{"url": url}, nil)
}

func (c *SubscriptionController) getStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	providerName := ctx.Query("provider")
	if providerName == "" {
		providerName = "xray"
	}
	status, err := c.subscriptionService.GetStatus(ctx.Request.Context(), providerName, id)
	if err != nil {
		jsonMsg(ctx, "failed to get subscription status", err)
		return
	}
	jsonObj(ctx, status, nil)
}

func (c *SubscriptionController) syncUsage(ctx *gin.Context) {
	id := ctx.Param("id")
	var req providerRequest
	_ = ctx.ShouldBindJSON(&req)
	if req.Provider == "" {
		req.Provider = "xray"
	}
	report, err := c.subscriptionService.SyncUsage(ctx.Request.Context(), req.Provider, id)
	if err != nil {
		jsonMsg(ctx, "failed to sync usage", err)
		return
	}
	jsonObj(ctx, report, nil)
}

func (c *SubscriptionController) list(ctx *gin.Context) {
	providerName := ctx.Query("provider")
	status := ctx.Query("status")

	providerId := 0
	if providerName != "" {
		rec, err := c.providerService.GetByName(providerName)
		if err == nil {
			providerId = rec.Id
		}
	}
	if pidStr := ctx.Query("providerId"); pidStr != "" {
		if pid, err := strconv.Atoi(pidStr); err == nil {
			providerId = pid
		}
	}

	subs, err := c.subscriptionService.ListExternal(providerId, status)
	if err != nil {
		jsonMsg(ctx, "failed to list subscriptions", err)
		return
	}
	jsonObj(ctx, subs, nil)
}
