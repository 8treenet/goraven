package agent

import (
	"context"
	"fmt"

	"goraven/core/iface"

	"github.com/cloudwego/eino/adk/middlewares/skill"
)

type systemSkillBackend struct {
	skills []iface.SkillInfo
}

func newSystemSkillBackend(provider iface.SystemSkillProvider) (skill.Backend, error) {
	skills, err := provider.SystemSkillList()
	if err != nil {
		return nil, err
	}
	return &systemSkillBackend{skills: skills}, nil
}

func (b *systemSkillBackend) List(ctx context.Context) ([]skill.FrontMatter, error) {
	result := make([]skill.FrontMatter, len(b.skills))
	for i, s := range b.skills {
		result[i] = skill.FrontMatter{
			Name:        s.Name,
			Description: s.Description,
		}
	}
	return result, nil
}

func (b *systemSkillBackend) Get(ctx context.Context, name string) (skill.Skill, error) {
	for _, s := range b.skills {
		if s.Name == name {
			return skill.Skill{
				FrontMatter: skill.FrontMatter{
					Name:        s.Name,
					Description: s.Description,
				},
				Content: s.Content,
			}, nil
		}
	}
	return skill.Skill{}, fmt.Errorf("skill %s not found", name)
}
