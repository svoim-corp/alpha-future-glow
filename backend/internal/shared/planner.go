package shared

import "context"

// PlanOwners — принадлежность планов зоны Planner. Реализует: разработчик B. Использует: Dashboard (B) при привязке виджета.
type PlanOwners interface {
	// PlanCompany — компания плана; "" — плана нет.
	PlanCompany(ctx context.Context, id string) (string, error)
}
