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

package application

import (
	"iter"
	"strings"
)

func MatchString(value, pattern string) bool {
	return strings.Contains(value, pattern)
}

func MatchStrings(values []string, pattern string) bool {
	for _, value := range values {
		if MatchString(value, pattern) {
			return true
		}
	}
	return false
}

type EntityFilter[T any] interface {
	Match(t T, query string) bool
}

type EntityFilterFunc[T any] func(t T, query string) bool

func (f EntityFilterFunc[T]) Match(t T, query string) bool {
	return f(t, query)
}

func Match[T any](in iter.Seq[T], query string, filter EntityFilter[T]) iter.Seq[T] {
	if len(query) == 0 {
		return in
	}
	return func(yield func(T) bool) {
		for t := range in {
			if filter.Match(t, query) {
				if !yield(t) {
					return
				}
				break
			}
		}
	}
}

func Limit[T any](in iter.Seq[T], limit int) iter.Seq[T] {
	if limit <= 0 {
		return in
	}
	remaining := limit
	return func(yield func(T) bool) {
		if remaining <= 0 {
			return
		}
		for t := range in {
			if !yield(t) {
				return
			}
			remaining--
			if remaining <= 0 {
				return
			}
		}
	}
}
