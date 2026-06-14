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

package mcp

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func addTaskTools(server *mcp.Server, provider domain.TaskProvider) {
	addSearchTasksTool(server, provider)
	addGetTaskTool(server, provider)
}

func addSearchTasksTool(server *mcp.Server, provider domain.TaskProvider) {
	tool := &mcp.Tool{
		Name:        "searchTasks",
		Description: "Searches for tasks using the given search parameters. A task summary including the task ID is returned for every found task. The task ID can be used to get the full task details (getTask).",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *SearchTasksParams) (*mcp.CallToolResult, any, error) {
		filter := domain.TaskFilter{
			StandardFilter: domain.StandardFilter{
				Query: params.Query,
				Limit: params.Limit,
			},
			Status:    (*domain.TaskStatus)(params.Status),
			DueAfter:  params.DueAfter,
			DueBefore: params.DueBefore,
		}
		tasks, err := provider.SearchTasks(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, toTaskSummaryOutputs(tasks), nil
	}
	mcp.AddTool(server, tool, handler)
}

func addGetTaskTool(server *mcp.Server, provider domain.TaskProvider) {
	tool := &mcp.Tool{
		Name:        "getTask",
		Description: "Gets the full task details for the given ID",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *GetTaskParams) (*mcp.CallToolResult, any, error) {
		task, err := provider.GetTask(ctx, params.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, toTaskOutput(task), nil
	}
	mcp.AddTool(server, tool, handler)
}

type SearchTasksParams struct {
	Query     *string    `json:"query,omitempty"      jsonschema:"Term to search for. All task attributes (Title, Description, Status) are matched against this term (substring match). As soon as one attribute matches, the task is included in the result. Leave empty to list all tasks."`
	Limit     *int       `json:"limit,omitempty"      jsonschema:"The maximum number of tasks to return. If no limit is given a provider specific one applies."`
	Status    *string    `json:"status,omitempty"     jsonschema:"Only return tasks with this status. Known status values are todo, in_progress, done. Leave empty to return all tasks regardless of status."`
	DueAfter  *time.Time `json:"due_after,omitempty"  jsonschema:"Only return tasks due at or after this time. Use RFC3339 format (e.g. 2026-06-21T00:00:00Z)."`
	DueBefore *time.Time `json:"due_before,omitempty" jsonschema:"Only return tasks due at or before this time. Use RFC3339 format (e.g. 2026-06-14T00:00:00Z)."`
}

type GetTaskParams struct {
	ID string `json:"id" jsonschema:"ID of the task to return."`
}

type TaskSummaryOutput struct {
	ID       string        `json:"id" jsonschema:"ID of the task."`
	Title    string        `json:"title" jsonschema:"The title of the task"`
	Status   string        `json:"status" jsonschema:"The status of the task (todo, in_progress, done)"`
	Priority string        `json:"priority" jsonschema:"The priority of the task (low, medium, high)"`
	DueAt    *TZTimeOutput `json:"due_at" jsonschema:"The due date of the task (RFC3339 format, timezone-aware). null if no due date is set."`
}

type TaskOutput struct {
	ID          string        `json:"id" jsonschema:"ID of the task."`
	Title       string        `json:"title" jsonschema:"The title of the task"`
	Description string        `json:"description" jsonschema:"The description of the task"`
	Status      string        `json:"status" jsonschema:"The status of the task (todo, in_progress, done)"`
	Priority    string        `json:"priority" jsonschema:"The priority of the task (low, medium, high)"`
	DueAt       *TZTimeOutput `json:"due_at" jsonschema:"The due date of the task (RFC3339 format, timezone-aware). null if no due date is set."`
	CompletedAt *TZTimeOutput `json:"completed_at" jsonschema:"The date the task has been completed (RFC3339 format, timezone-aware). null if not yet completed."`
	CreatedAt   time.Time     `json:"created_at" jsonschema:"The date the task has been created (RFC3339 format)."`
	UpdatedAt   time.Time     `json:"updated_at" jsonschema:"The last time the task was updated (RFC3339 format)."`
}

func toTaskSummaryOutputs(tasks []*domain.Task) []*TaskSummaryOutput {
	outputs := make([]*TaskSummaryOutput, 0, len(tasks))
	for _, task := range tasks {
		output := &TaskSummaryOutput{
			ID:       task.ID,
			Title:    task.Title,
			Status:   string(task.Status),
			Priority: string(task.Priority),
			DueAt:    toTZTimeOutputPtr(task.DueAt),
		}
		outputs = append(outputs, output)
	}
	return outputs
}

func toTaskOutput(task *domain.Task) *TaskOutput {
	output := &TaskOutput{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		DueAt:       toTZTimeOutputPtr(task.DueAt),
		CompletedAt: toTZTimeOutputPtr(task.CompletedAt),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
	return output
}
