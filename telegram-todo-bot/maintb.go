package main

import (
	"sync"
)

type Task struct {
	ID    int
	Title string
	Done  bool
}

type Storage struct {
	mu     sync.RWMutex
	tasks  map[int]Task
	nextID int
}

func NewStorage() *Storage {
	return &Storage{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

func (s *Storage) add(title string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++
	s.tasks[id] = Task{ID: id, Title: title, Done: false}
	return id
}

func (s *Storage) list() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].ID > tasks[j].ID {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
	return tasks
}

func (s *Storage) done(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return false
	}
	task.Done = true
	s.tasks[id] = task
	return true
}

func (s *Storage) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.tasks[id]
	if !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}
