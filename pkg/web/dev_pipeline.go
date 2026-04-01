package web

import (
	"context"

	"github.com/sipeed/makoclaw/pkg/devmemory"
	"github.com/sipeed/makoclaw/pkg/logger"
)

type DevRequest struct {
	UserUUID        string
	Prompt          string
	PromptInjection string
}

type DevResponse struct {
	UserUUID  string
	SessionID string
}

type PreMiddlewareFunc func(ctx context.Context, req *DevRequest) (*DevRequest, error)

type PostMiddlewareFunc func(ctx context.Context, resp *DevResponse) error

type DevPipeline struct {
	pre  []PreMiddlewareFunc
	post []PostMiddlewareFunc
}

func NewDevPipeline() *DevPipeline {
	return &DevPipeline{}
}

func (p *DevPipeline) Use(fn PreMiddlewareFunc) {
	if p == nil || fn == nil {
		return
	}
	p.pre = append(p.pre, fn)
}

func (p *DevPipeline) UsePost(fn PostMiddlewareFunc) {
	if p == nil || fn == nil {
		return
	}
	p.post = append(p.post, fn)
}

func (p *DevPipeline) RunPre(ctx context.Context, req *DevRequest) (*DevRequest, error) {
	if p == nil || len(p.pre) == 0 {
		return req, nil
	}

	current := req
	for _, middleware := range p.pre {
		next, err := middleware(ctx, current)
		if err != nil {
			return nil, err
		}
		current = next
	}

	return current, nil
}

func (p *DevPipeline) RunPost(ctx context.Context, resp *DevResponse) error {
	if p == nil || len(p.post) == 0 {
		return nil
	}

	for _, middleware := range p.post {
		if err := middleware(ctx, resp); err != nil {
			logger.WarnCF("web", "dev pipeline post middleware failed", map[string]interface{}{"error": err.Error(), "user_uuid": resp.UserUUID, "session_id": resp.SessionID})
		}
	}

	return nil
}

func MemoryInjectionMiddleware(getDevMem func(userUUID string) (*devmemory.Store, error)) PreMiddlewareFunc {
	return func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
		if req == nil || getDevMem == nil {
			return req, nil
		}

		devMem, err := getDevMem(req.UserUUID)
		if err != nil || devMem == nil {
			return req, nil
		}

		injected, err := devMem.Inject(ctx, req.Prompt, 5)
		if err != nil || injected == "" {
			return req, nil
		}

		req.PromptInjection = injected
		return req, nil
	}
}
