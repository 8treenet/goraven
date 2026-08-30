package dependency

import (
	"goraven/backend/po"
	"time"
)

type Chat interface {
	AutomationTask(task *po.AutomationTask, startedAt time.Time)
}
