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

package msgraph

import (
	"context"
	"fmt"
	"slices"
	"time"

	kiota "github.com/microsoft/kiota-abstractions-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchTasks(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, error) {
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	listID, err := p.defaultTaskListID(ctx, client)
	if err != nil {
		return nil, err
	}
	requestConfig := p.taskFilterRequestConfig(filter)
	response, err := client.Me().Todo().Lists().ByTodoTaskListId(listID).Tasks().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("search tasks Graph API failure (cause: %w)", err)
	}
	tasks := make([]*domain.Task, 0)
	for _, responseItem := range response.GetValue() {
		task := p.taskFromResponse(responseItem)
		if !task.Empty() {
			tasks = append(tasks, task)
		}
	}
	slices.SortFunc(tasks, application.TaskSortFunc)
	return tasks, nil
}

func (p *Provider) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	listID, err := p.defaultTaskListID(ctx, client)
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Todo().Lists().ByTodoTaskListId(listID).Tasks().ByTodoTaskId(id).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get task Graph API failure (cause: %w)", err)
	}
	task := p.taskFromResponse(response)
	return task, nil
}

func (p *Provider) CreateTask(ctx context.Context, create domain.TaskCreate) (*domain.Task, error) {
	request := models.NewTodoTask()
	request.SetTitle(&create.Title)
	if create.Description != nil {
		body := models.NewItemBody()
		body.SetContentType(bodyTypePtr(models.TEXT_BODYTYPE))
		body.SetContent(create.Description)
		request.SetBody(body)
	}
	if create.Status != nil {
		switch *create.Status {
		case domain.StatusTodo:
			request.SetStatus(taskStatusPtr(models.NOTSTARTED_TASKSTATUS))
		case domain.StatusInProgress:
			request.SetStatus(taskStatusPtr(models.INPROGRESS_TASKSTATUS))
		case domain.StatusDone:
			request.SetStatus(taskStatusPtr(models.COMPLETED_TASKSTATUS))
		}
	}
	if create.Priority != nil {
		switch *create.Priority {
		case domain.PriorityLow:
			request.SetImportance(importancePtr(models.LOW_IMPORTANCE))
		case domain.PriorityMedium:
			request.SetImportance(importancePtr(models.NORMAL_IMPORTANCE))
		case domain.PriorityHigh:
			request.SetImportance(importancePtr(models.HIGH_IMPORTANCE))
		}
	}
	if create.DueAt != nil {
		dueDateTime := models.NewDateTimeTimeZone()
		dateTime, timezone := marshalTZTime(*create.DueAt)
		dueDateTime.SetDateTime(dateTime)
		dueDateTime.SetTimeZone(timezone)
		request.SetDueDateTime(dueDateTime)
	}
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	listID, err := p.defaultTaskListID(ctx, client)
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Todo().Lists().ByTodoTaskListId(listID).Tasks().Post(ctx, request, nil)
	if err != nil {
		return nil, fmt.Errorf("create task Graph API failure (cause: %w)", err)
	}
	task := p.taskFromResponse(response)
	return task, nil
}

func (p *Provider) UpdateTask(ctx context.Context, id string, update domain.TaskUpdate) (*domain.Task, error) {
	request := models.NewTodoTask()
	request.SetTitle(update.Title)
	if update.Description != nil {
		body := models.NewItemBody()
		body.SetContentType(bodyTypePtr(models.TEXT_BODYTYPE))
		body.SetContent(update.Description)
		request.SetBody(body)
	}
	if update.Status != nil {
		switch *update.Status {
		case domain.StatusTodo:
			request.SetStatus(taskStatusPtr(models.NOTSTARTED_TASKSTATUS))
		case domain.StatusInProgress:
			request.SetStatus(taskStatusPtr(models.INPROGRESS_TASKSTATUS))
		case domain.StatusDone:
			request.SetStatus(taskStatusPtr(models.COMPLETED_TASKSTATUS))
		}
	}
	if update.Priority != nil {
		switch *update.Priority {
		case domain.PriorityLow:
			request.SetImportance(importancePtr(models.LOW_IMPORTANCE))
		case domain.PriorityMedium:
			request.SetImportance(importancePtr(models.NORMAL_IMPORTANCE))
		case domain.PriorityHigh:
			request.SetImportance(importancePtr(models.HIGH_IMPORTANCE))
		}
	}
	if update.DueAt != nil {
		dueDateTime := models.NewDateTimeTimeZone()
		dateTime, timezone := marshalTZTime(*update.DueAt)
		dueDateTime.SetDateTime(dateTime)
		dueDateTime.SetTimeZone(timezone)
		request.SetDueDateTime(dueDateTime)
	}
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	listID, err := p.defaultTaskListID(ctx, client)
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Todo().Lists().ByTodoTaskListId(listID).Tasks().ByTodoTaskId(id).Patch(ctx, request, nil)
	if err != nil {
		return nil, fmt.Errorf("update task Graph API failure (cause: %w)", err)
	}
	task := p.taskFromResponse(response)
	return task, nil
}

func (p *Provider) defaultTaskListID(ctx context.Context, client *msgraphsdk.GraphServiceClient) (string, error) {
	response, err := client.Me().Todo().Lists().Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("list task lists Graph API failure (cause: %w)", err)
	}
	for _, list := range response.GetValue() {
		if list.GetWellknownListName().String() == "defaultList" {
			id := list.GetId()
			if id != nil && *id != "" {
				return *id, nil
			}
			break
		}
	}
	return "", fmt.Errorf("no default task list found")
}

func (p *Provider) taskFromResponse(model models.TodoTaskable) *domain.Task {
	body := model.GetBody()
	content := ""
	if body != nil {
		content = *body.GetContent()
	}
	return &domain.Task{
		ID:          ptrString(model.GetId()),
		Title:       ptrString(model.GetTitle()),
		Description: content,
		Status:      p.taskStatusFromResponse(model),
		Priority:    p.taskPriorityFromResponse(model),
		DueAt:       p.taskDateTimeFromResponse(model.GetDueDateTime()),
		CompletedAt: p.taskDateTimeFromResponse(model.GetCompletedDateTime()),
		CreatedAt:   ptrTime(model.GetCreatedDateTime()),
		UpdatedAt:   ptrTime(model.GetLastModifiedDateTime()),
	}
}

func (p *Provider) taskDateTimeFromResponse(model models.DateTimeTimeZoneable) *domain.TZTime {
	if model == nil {
		return nil
	}
	parsed := unmarshalTZTime(model.GetDateTime(), model.GetTimeZone(), p.cfg.MSGraph.DefaultTimeLocation.Location)
	if parsed.Empty() {
		return nil
	}
	return &parsed
}

func (p *Provider) taskStatusFromResponse(model models.TodoTaskable) domain.TaskStatus {
	status := model.GetStatus()
	if status == nil {
		return domain.StatusTodo
	}
	switch *status {
	case models.NOTSTARTED_TASKSTATUS:
		return domain.StatusTodo
	case models.COMPLETED_TASKSTATUS:
		return domain.StatusDone
	default:
		return domain.StatusInProgress
	}
}

func (p *Provider) taskPriorityFromResponse(model models.TodoTaskable) domain.TaskPriority {
	importance := model.GetImportance()
	if importance == nil {
		return domain.PriorityLow
	}
	switch *importance {
	case models.HIGH_IMPORTANCE:
		return domain.PriorityHigh
	case models.NORMAL_IMPORTANCE:
		return domain.PriorityMedium
	default:
		return domain.PriorityLow
	}
}

func (p *Provider) taskFilterRequestConfig(filter domain.TaskFilter) *users.ItemTodoListsItemTasksRequestBuilderGetRequestConfiguration {
	search, limit := standardFilterPtr(filter.StandardFilter)
	nowUTC := time.Now().UTC()
	var dueAfter string
	if filter.DueAfter != nil && !filter.DueAfter.IsZero() {
		dueAfter = filter.DueAfter.UTC().Format(time.RFC3339)
	} else {
		dueAfter = nowUTC.Format(time.RFC3339)
	}
	var dueBefore string
	if filter.DueBefore != nil && !filter.DueBefore.IsZero() {
		dueBefore = filter.DueBefore.UTC().Format(time.RFC3339)
	} else {
		dueBefore = nowUTC.Add(7 * 24 * time.Hour).Format(time.RFC3339)
	}
	filterParam := fmt.Sprintf("(dueDateTime/dateTime ge '%s') and (dueDateTime/dateTime le '%s')", dueAfter, dueBefore)
	headers := &kiota.RequestHeaders{}
	headers.Add("ConsistencyLevel", "eventual")
	headers.Add("Prefer", "outlook.body-content-type=\"text\"")
	requestConfig := &users.ItemTodoListsItemTasksRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemTodoListsItemTasksRequestBuilderGetQueryParameters{
			Search: search,
			Filter: &filterParam,
			Top:    limit,
			Count:  boolPtr(true),
		},
		Headers: headers,
	}
	return requestConfig
}
