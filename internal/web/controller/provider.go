package controller

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/provider"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"

	"github.com/gin-gonic/gin"
)

type ProviderController struct {
	BaseController
	providerService service.ProviderService
	registry        *provider.Registry
}

func NewProviderController(g *gin.RouterGroup, registry *provider.Registry) *ProviderController {
	c := &ProviderController{registry: registry}
	c.initRouter(g)
	return c
}

func (c *ProviderController) initRouter(g *gin.RouterGroup) {
	g.GET("/list", c.list)
	g.GET("/get/:id", c.get)
	g.POST("/add", c.add)
	g.POST("/update/:id", c.update)
	g.POST("/del/:id", c.del)
	g.POST("/test/:id", c.test)
}

func (c *ProviderController) list(ctx *gin.Context) {
	providers, err := c.providerService.GetAll()
	if err != nil {
		jsonMsg(ctx, "failed to list providers", err)
		return
	}
	jsonObj(ctx, providers, nil)
}

func (c *ProviderController) get(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		jsonMsg(ctx, "invalid provider id", err)
		return
	}
	p, err := c.providerService.GetById(id)
	if err != nil {
		jsonMsg(ctx, "provider not found", err)
		return
	}
	jsonObj(ctx, p, nil)
}

func (c *ProviderController) add(ctx *gin.Context) {
	var rec model.ProviderRecord
	if err := ctx.ShouldBindJSON(&rec); err != nil {
		jsonMsg(ctx, "invalid request body", err)
		return
	}
	if rec.Name == "" {
		jsonMsg(ctx, "provider name is required", errors.New("name is empty"))
		return
	}
	if rec.Type == "" {
		rec.Type = string(provider.ProviderTypeExternal)
	}
	if rec.Config == "" {
		rec.Config = "{}"
	}
	rec.IsEnabled = true
	if err := c.providerService.Create(&rec); err != nil {
		jsonMsg(ctx, "failed to create provider", err)
		return
	}
	jsonObj(ctx, rec, nil)
}

func (c *ProviderController) update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		jsonMsg(ctx, "invalid provider id", err)
		return
	}
	existing, err := c.providerService.GetById(id)
	if err != nil {
		jsonMsg(ctx, "provider not found", err)
		return
	}
	var updates model.ProviderRecord
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		jsonMsg(ctx, "invalid request body", err)
		return
	}
	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.Type != "" {
		existing.Type = updates.Type
	}
	if updates.Config != "" {
		existing.Config = updates.Config
	}
	existing.IsEnabled = updates.IsEnabled
	if err := c.providerService.Update(existing); err != nil {
		jsonMsg(ctx, "failed to update provider", err)
		return
	}
	jsonObj(ctx, existing, nil)
}

func (c *ProviderController) del(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		jsonMsg(ctx, "invalid provider id", err)
		return
	}
	if err := c.providerService.Delete(id); err != nil {
		jsonMsg(ctx, "failed to delete provider", err)
		return
	}
	jsonMsg(ctx, "provider deleted", nil)
}

func (c *ProviderController) test(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		jsonMsg(ctx, "invalid provider id", err)
		return
	}
	rec, err := c.providerService.GetById(id)
	if err != nil {
		jsonMsg(ctx, "provider not found", err)
		return
	}
	p, ok := c.registry.Get(rec.Name)
	if !ok {
		jsonObj(ctx, map[string]any{"healthy": false, "reason": "provider not registered"}, nil)
		return
	}
	tctx, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()
	healthy := p.IsHealthy(tctx)
	jsonObj(ctx, map[string]any{"healthy": healthy}, nil)
}
