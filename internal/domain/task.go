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

package domain

import (
	"context"
	"strings"
	"time"
)

type Task struct {
	ID          string
	Title       string
	Description string
	Status      TaskStatus
	Priority    TaskPriority
	DueAt       *TZTime
	CompletedAt *TZTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t *Task) String() string {
	buffer := &strings.Builder{}
	buffer.WriteString(string(t.Priority))
	buffer.WriteRune(' ')
	buffer.WriteString(string(t.Status))
	buffer.WriteRune(' ')
	buffer.WriteString(t.Title)
	if t.DueAt != nil {
		buffer.WriteRune(' ')
		buffer.WriteString(t.DueAt.String())
	}
	return buffer.String()
}

func (t *Task) Empty() bool {
	return t.ID == ""
}

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

type TaskFilter struct {
	StandardFilter
	Status    *TaskStatus
	DueAfter  *time.Time
	DueBefore *time.Time
}

// TaskCreate beschreibt die Felder zum Anlegen eines neuen Tasks.
// Title ist Pflicht, alle anderen Felder optional.
type TaskCreate struct {
	Title       string // Pflicht — jeder Task braucht einen Titel
	Description *string
	Status      *TaskStatus
	Priority    *TaskPriority
	DueAt       *TZTime
}

// TaskUpdate beschreibt, welche Felder eines Tasks geändert werden sollen.
// Nur non-nil Felder werden angewandt (PATCH-Semantik).
// DueAt mit zero DateTime (IsZero()==true) löscht das Fälligkeitsdatum.
type TaskUpdate struct {
	Title       *string
	Description *string
	Status      *TaskStatus
	Priority    *TaskPriority
	DueAt       *TZTime // nil = nicht anfassen; &TZTime{} = löschen; sonst setzen
}

// TaskProvider ist das Lese-Interface — alle Adapter implementieren es.
type TaskProvider interface {
	SearchTasks(ctx context.Context, filter TaskFilter) ([]*Task, error)
	GetTask(ctx context.Context, id string) (*Task, error)
}

// TaskWriteProvider erweitert TaskProvider um Schreib-Operationen.
// Wird nur registriert, wenn ProviderCapabilities.AccessMode == ReadWrite.
type TaskWriteProvider interface {
	TaskProvider
	CreateTask(ctx context.Context, create TaskCreate) (*Task, error)
	UpdateTask(ctx context.Context, id string, update TaskUpdate) (*Task, error)
}
