package middleware

import (
	"context"
	"fmt"

	"goraven/core/iface"

	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/adk/middlewares/skill"
)

type SkillBackendConfig struct {
	Backend filesystem.Backend
	BaseDir string
}

type CombinedBackend struct {
	userBackend   skill.Backend
	sysBackend    skill.Backend
	skillAccessor iface.SkillAccessor
}

func NewCombinedBackend(userBackend, sysBackend skill.Backend, filter iface.SkillAccessor) *CombinedBackend {
	return &CombinedBackend{
		userBackend:   userBackend,
		sysBackend:    sysBackend,
		skillAccessor: filter,
	}
}

func NewCombinedSkillBackend(ctx context.Context, userCfg, systemCfg *SkillBackendConfig, accessor iface.SkillAccessor) (skill.Backend, error) {
	var (
		userBackend skill.Backend
		sysBackend  skill.Backend
	)

	if userCfg != nil {
		b, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
			Backend: userCfg.Backend,
			BaseDir: userCfg.BaseDir,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create user skill backend: %w", err)
		}
		userBackend = b
	}

	if systemCfg != nil {
		b, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
			Backend: systemCfg.Backend,
			BaseDir: systemCfg.BaseDir,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create system skill backend: %w", err)
		}
		sysBackend = b
	}

	return NewCombinedBackend(userBackend, sysBackend, accessor), nil
}

func NewCombinedSkillBackendWithSystemBackend(ctx context.Context, userCfg *SkillBackendConfig, systemBackend skill.Backend, accessor iface.SkillAccessor) (skill.Backend, error) {
	var userBackend skill.Backend

	if userCfg != nil {
		b, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
			Backend: userCfg.Backend,
			BaseDir: userCfg.BaseDir,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create user skill backend: %w", err)
		}
		userBackend = b
	}

	return NewCombinedBackend(userBackend, systemBackend, accessor), nil
}

func (cb *CombinedBackend) List(ctx context.Context) ([]skill.FrontMatter, error) {
	var result []skill.FrontMatter
	if cb.sysBackend != nil {
		sysSkills, err := cb.sysBackend.List(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, sysSkills...)
	}

	if cb.userBackend != nil {
		userSkills, err := cb.userBackend.List(ctx)
		if err != nil {
			return nil, err
		}
		var selectUserSkills []skill.FrontMatter
		for _, s := range userSkills {
			if cb.skillAccessor != nil && !cb.skillAccessor.IsSkillSelected(s.Name) {
				//如果设置了skillFilter，并且未勾选
				continue
			}
			if len(selectUserSkills) >= 60 {
				break
			}
			selectUserSkills = append(selectUserSkills, s)
		}
		result = append(result, selectUserSkills...)
	}
	return result, nil
}

func (cb *CombinedBackend) Get(ctx context.Context, name string) (skill.Skill, error) {
	if cb.skillAccessor != nil {
		cb.skillAccessor.AddToolDailyStats(name)
	}

	if cb.sysBackend != nil {
		s, err := cb.sysBackend.Get(ctx, name)
		if err == nil {
			return s, nil
		}
	}
	if cb.userBackend != nil {
		return cb.userBackend.Get(ctx, name)
	}
	return skill.Skill{}, fmt.Errorf("skill %s not found", name)
}
