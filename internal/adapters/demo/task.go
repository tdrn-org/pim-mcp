/*
 * Copyright 2026 Holger de Carne
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package demo

import (
	"context"
	"iter"
	"maps"
	"slices"
	"time"

	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchTasks(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, error) {
	tasks := slices.Collect(maps.Values(taskData))
	slices.SortFunc(tasks, application.TaskSortFunc)
	taskList := slices.Values(tasks)
	taskList = application.Match(taskList, filter.Query, application.EntityFilterFunc[*domain.Task](application.TaskFilterFunc))
	taskList = filterByStatus(taskList, filter.Status)
	taskList = filterByDueAfter(taskList, filter.DueAfter)
	taskList = filterByDueBefore(taskList, filter.DueBefore)
	taskList = application.Limit(taskList, filter.Limit)
	return slices.Collect(taskList), nil
}

func (p *Provider) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	task, ok := taskData[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return task, nil
}

func filterByStatus(in iter.Seq[*domain.Task], status *domain.TaskStatus) iter.Seq[*domain.Task] {
	if status == nil || *status == "" {
		return in
	}
	return func(yield func(*domain.Task) bool) {
		for task := range in {
			if task.Status == *status {
				if !yield(task) {
					return
				}
			}
		}
	}
}

func filterByDueAfter(in iter.Seq[*domain.Task], dueAfter *time.Time) iter.Seq[*domain.Task] {
	if dueAfter == nil || dueAfter.IsZero() {
		return in
	}
	return func(yield func(*domain.Task) bool) {
		for task := range in {
			if task.DueAt != nil && !task.DueAt.DateTime.Before(*dueAfter) {
				if !yield(task) {
					return
				}
			}
		}
	}
}

func filterByDueBefore(in iter.Seq[*domain.Task], dueBefore *time.Time) iter.Seq[*domain.Task] {
	if dueBefore == nil || dueBefore.IsZero() {
		return in
	}
	return func(yield func(*domain.Task) bool) {
		for task := range in {
			if task.DueAt != nil && !task.DueAt.DateTime.After(*dueBefore) {
				if !yield(task) {
					return
				}
			}
		}
	}
}

var taskData map[string]*domain.Task = map[string]*domain.Task{
	"t1": {
		ID:          "t1",
		Title:       "Update authentication module to v4.2.1",
		Description: "Apply the security patch for CVE-2026-12345. Update auth-service and api-gateway dependencies. Run integration tests after deployment.",
		Status:      domain.StatusTodo,
		Priority:    domain.PriorityHigh,
		DueAt:       toTZTimePtr(time.Date(2026, 6, 17, 18, 0, 0, 0, time.UTC), "Europe/Berlin"),
		CompletedAt: nil,
		CreatedAt:   time.Date(2026, 6, 15, 7, 5, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 15, 7, 5, 0, 0, time.UTC),
	},
	"t2": {
		ID:          "t2",
		Title:       "Prepare Q2 roadmap presentation",
		Description: "Create slides for the quarterly roadmap review. Include progress metrics, milestone status, and Q3 proposals. Share draft with team leads by Wednesday.",
		Status:      domain.StatusInProgress,
		Priority:    domain.PriorityHigh,
		DueAt:       toTZTimePtr(time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC), "Europe/Berlin"),
		CompletedAt: nil,
		CreatedAt:   time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 14, 11, 30, 0, 0, time.UTC),
	},
	"t3": {
		ID:          "t3",
		Title:       "Review budget proposal from Finance",
		Description: "Go through the Q3 budget proposal document. Check infrastructure cost projections and flag any discrepancies. Provide feedback by Friday.",
		Status:      domain.StatusTodo,
		Priority:    domain.PriorityMedium,
		DueAt:       toTZTimePtr(time.Date(2026, 6, 19, 17, 0, 0, 0, time.UTC), "Europe/Berlin"),
		CompletedAt: nil,
		CreatedAt:   time.Date(2026, 6, 12, 14, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 12, 14, 0, 0, 0, time.UTC),
	},
	"t4": {
		ID:          "t4",
		Title:       "Book travel for DevCon 2026",
		Description: "Confirm flight and hotel reservations. Submit travel request form to HR. Check visa requirements if applicable.",
		Status:      domain.StatusDone,
		Priority:    domain.PriorityMedium,
		DueAt:       toTZTimePtr(time.Date(2026, 6, 10, 17, 0, 0, 0, time.UTC), "Europe/Berlin"),
		CompletedAt: toTZTimePtr(time.Date(2026, 6, 9, 15, 30, 0, 0, time.UTC), "Europe/Berlin"),
		CreatedAt:   time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 9, 15, 30, 0, 0, time.UTC),
	},
	"t5": {
		ID:          "t5",
		Title:       "Organize team building event",
		Description: "Plan a team building activity for the engineering department. Research options (escape room, cooking class, outdoor activity). Get budget approval from HR. Send out poll for date preferences.",
		Status:      domain.StatusTodo,
		Priority:    domain.PriorityLow,
		DueAt:       toTZTimePtr(time.Date(2026, 7, 1, 17, 0, 0, 0, time.UTC), "Europe/Berlin"),
		CompletedAt: nil,
		CreatedAt:   time.Date(2026, 6, 8, 13, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 8, 13, 0, 0, 0, time.UTC),
	},
}

func toTZTimePtr(dateTime time.Time, timezone string) *domain.TZTime {
	tzt := domain.NewTZTime(dateTime, timezone)
	return &tzt
}
