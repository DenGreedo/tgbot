package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func formatTasksList(tasks []Task) string {
	if len(tasks) == 0 {
		return "Список задач пуст."
	}
	var sb strings.Builder
	sb.WriteString("Ваши задачи:\n")
	for _, t := range tasks {
		status := "⬜"
		if t.Done {
			status = "✅"
		}
		fmt.Fprintf(&sb, "%d. %s %s\n", t.ID, status, t.Title)
	}
	return sb.String()
}

func handleCommand(bot *tgbotapi.BotAPI, storage *Storage, chatID int64, command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "/start":
		sendMessage(bot, chatID, "Привет! Я бот для управления задачами.\nИспользуй /help для списка команд.")

	case "/help":
		helpText := "Доступные команды:\n" +
			"/add <текст> — добавить задачу\n" +
			"/list — показать все задачи\n" +
			"/done <id> — отметить задачу выполненной\n" +
			"/delete <id> — удалить задачу"
		sendMessage(bot, chatID, helpText)

	case "/add":
		if len(parts) < 2 {
			sendMessage(bot, chatID, "Использование: /add <текст задачи>")
			return
		}
		title := strings.Join(parts[1:], " ")
		id := storage.add(title)
		sendMessage(bot, chatID, "Задача добавлена с ID: "+strconv.Itoa(id))

	case "/list":
		tasks := storage.list()
		sendMessage(bot, chatID, formatTasksList(tasks))

	case "/done":
		if len(parts) != 2 {
			sendMessage(bot, chatID, "Использование: /done <id>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			sendMessage(bot, chatID, "ID должен быть числом")
			return
		}
		if storage.done(id) {
			sendMessage(bot, chatID, "Задача #"+strconv.Itoa(id)+" отмечена выполненной ✅")
		} else {
			sendMessage(bot, chatID, "Задача с таким ID не найдена")
		}

	case "/delete":
		if len(parts) != 2 {
			sendMessage(bot, chatID, "Использование: /delete <id>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			sendMessage(bot, chatID, "ID должен быть числом")
			return
		}
		if storage.delete(id) {
			sendMessage(bot, chatID, "Задача #"+strconv.Itoa(id)+" удалена")
		} else {
			sendMessage(bot, chatID, "Задача с таким ID не найдена")
		}

	default:
		sendMessage(bot, chatID, "Неизвестная команда. Напиши /help для списка.")
	}
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("Не задан TELEGRAM_BOT_TOKEN. Получите токен у @BotFather и установите переменную окружения.")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Ошибка подключения к Telegram API: %v", err)
	}

	log.Printf("Бот запущен как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	storage := NewStorage()

	for update := range updates {
		if update.Message == nil {
			continue
		}
		go handleCommand(bot, storage, update.Message.Chat.ID, update.Message.Text)
	}
}
