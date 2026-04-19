package handler_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"task-5/internal/handler"
	"task-5/internal/model"

	"github.com/go-playground/validator"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) CreateChat(ctx context.Context, chat *model.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockChatService) DeleteChat(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChatService) CreateMessage(ctx context.Context, msg *model.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockChatService) GetChatWithMessages(ctx context.Context, id uint, limit int) (*model.Chat, error) {
	args := m.Called(ctx, id, limit)

	var chat *model.Chat
	if args.Get(0) != nil {
		chat = args.Get(0).(*model.Chat)
	}

	return chat, args.Error(1)
}

type testEnv struct {
	service   *MockChatService
	logger    *slog.Logger
	validator *validator.Validate
}

func setupTestEnv(mockSetup func(*MockChatService)) testEnv {
	service := new(MockChatService)
	if mockSetup != nil {
		mockSetup(service)
	}

	return testEnv{
		service:   service,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		validator: validator.New(),
	}
}

func TestChatHandler_CreateChat(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func(m *MockChatService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Broken JSON",
			body:           `{"title": "Chat title"`,
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty chat title",
			body:           `{"title": ""}`,
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Chat title too long",
			body:           fmt.Sprintf(`{"title": "%s"}`, strings.Repeat("a", 201)),
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Chat title consists entirely of spaces",
			body:           `{"title": "    "}`,
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "Chat duplicate",
			body: `{"title": "Chat title"}`,
			mockSetup: func(m *MockChatService) {
				m.On(
					"CreateChat",
					mock.Anything,
					mock.MatchedBy(func(chat *model.Chat) bool {
						return chat.Title == "Chat title"
					})).Return(model.ErrAlreadyExists).Once()
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Success",
			body: `{"title": "Chat title"}`,
			mockSetup: func(m *MockChatService) {
				m.On(
					"CreateChat",
					mock.Anything,
					mock.MatchedBy(func(chat *model.Chat) bool {
						return chat.Title == "Chat title"
					}),
				).Run(func(args mock.Arguments) {
					chat := args.Get(1).(*model.Chat)

					chat.ID = 1
					chat.CreatedAt = time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
				}).Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1, "title":"Chat title", "created_at":"2026-04-20T00:00:00Z"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			env := setupTestEnv(tc.mockSetup)
			h := handler.NewChatHandler(env.service, env.validator, env.logger)

			mux := http.NewServeMux()
			mux.HandleFunc("POST /chat", h.CreateChat)

			path := "/chat"
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedBody != "" {
				require.JSONEq(t, tc.expectedBody, rec.Body.String())
			}

			env.service.AssertExpectations(t)
		})
	}
}

func TestChatHandler_DeleteChat(t *testing.T) {
	tests := []struct {
		name           string
		chatIDPath     string
		mockSetup      func(m *MockChatService)
		expectedStatus int
	}{
		{
			name:           "Invalid ID type",
			chatIDPath:     "abc",
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Negative ID",
			chatIDPath:     "-1",
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Not found",
			chatIDPath: "2",
			mockSetup: func(m *MockChatService) {
				m.On("DeleteChat", mock.Anything, uint(2)).Return(model.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "Success",
			chatIDPath: "1",
			mockSetup: func(m *MockChatService) {
				m.On("DeleteChat", mock.Anything, uint(1)).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			env := setupTestEnv(tc.mockSetup)
			h := handler.NewChatHandler(env.service, env.validator, env.logger)

			mux := http.NewServeMux()
			mux.HandleFunc("DELETE /chat/{chat_id}", h.DeleteChat)

			path := "/chat/" + tc.chatIDPath
			req := httptest.NewRequest(http.MethodDelete, path, nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)
			env.service.AssertExpectations(t)
		})
	}
}

func TestChatHandler_CreateMessage(t *testing.T) {
	tests := []struct {
		name           string
		chatIDPath     string
		body           string
		mockSetup      func(m *MockChatService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Invalid ID type",
			chatIDPath:     "abc",
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Negative ID",
			chatIDPath:     "-1",
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty message text",
			chatIDPath:     "1",
			body:           `{"text": ""}`,
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Broken JSON",
			chatIDPath:     "1",
			body:           `{"text": ""`,
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Message text too long",
			chatIDPath:     "1",
			body:           fmt.Sprintf(`{"text": "%s"}`, strings.Repeat("a", 5001)),
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Message text consists entirely of spaces",
			chatIDPath:     "1",
			body:           `{"text": "    "}`,
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "Not found",
			chatIDPath: "2",
			body:       `{"text": "Message text"}`,
			mockSetup: func(m *MockChatService) {
				m.On("CreateMessage", mock.Anything, mock.MatchedBy(func(msg *model.Message) bool {
					return msg.ChatID == 2 && msg.Text == "Message text"
				})).Return(model.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "Success",
			chatIDPath: "1",
			body:       `{"text": "Message text"}`,
			mockSetup: func(m *MockChatService) {
				m.On("CreateMessage", mock.Anything, mock.MatchedBy(func(msg *model.Message) bool {
					return msg.ChatID == 1 && msg.Text == "Message text"
				})).Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			env := setupTestEnv(tc.mockSetup)
			h := handler.NewChatHandler(env.service, env.validator, env.logger)

			mux := http.NewServeMux()
			mux.HandleFunc("POST /chat/{chat_id}/message", h.CreateMessage)

			path := "/chat/" + tc.chatIDPath + "/message"
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)
			env.service.AssertExpectations(t)
		})
	}
}

func TestChatHandler_GetAllMessages(t *testing.T) {
	tests := []struct {
		name           string
		chatIDPath     string
		limitQuery     string
		mockSetup      func(m *MockChatService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Invalid ID type",
			chatIDPath:     "abc",
			limitQuery:     "limit=20",
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Negative ID",
			chatIDPath:     "-1",
			limitQuery:     "limit=20",
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Not Found",
			chatIDPath: "2",
			limitQuery: "limit=20",
			mockSetup: func(m *MockChatService) {
				m.On("GetChatWithMessages", mock.Anything, uint(2), 20).Return(nil, model.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "Success",
			chatIDPath: "1",
			limitQuery: "limit=20",
			mockSetup: func(m *MockChatService) {
				m.On("GetChatWithMessages", mock.Anything, uint(1), 20).Return(&model.Chat{
					ID:        1,
					Title:     "Chat title",
					Messages:  []model.Message{},
					CreatedAt: time.Now(),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Without limit",
			chatIDPath: "1",
			limitQuery: "",
			mockSetup: func(m *MockChatService) {
				m.On("GetChatWithMessages", mock.Anything, uint(1), 5).Return(&model.Chat{
					ID:        1,
					Title:     "Chat title",
					Messages:  []model.Message{},
					CreatedAt: time.Now(),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Limit too big",
			chatIDPath: "1",
			limitQuery: "limit=100",
			mockSetup: func(m *MockChatService) {
				m.On("GetChatWithMessages", mock.Anything, uint(1), 20).Return(&model.Chat{
					ID:        1,
					Title:     "Chat title",
					Messages:  []model.Message{},
					CreatedAt: time.Now(),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Limit too small",
			chatIDPath: "1",
			limitQuery: "limit=-5",
			mockSetup: func(m *MockChatService) {
				m.On("GetChatWithMessages", mock.Anything, uint(1), 5).Return(&model.Chat{
					ID:        1,
					Title:     "Chat title",
					Messages:  []model.Message{},
					CreatedAt: time.Now(),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Zero limit",
			chatIDPath: "1",
			limitQuery: "limit=0",
			mockSetup: func(m *MockChatService) {
				m.On("GetChatWithMessages", mock.Anything, uint(1), 5).Return(&model.Chat{
					ID:        1,
					Title:     "Chat title",
					Messages:  []model.Message{},
					CreatedAt: time.Now(),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			env := setupTestEnv(tc.mockSetup)
			h := handler.NewChatHandler(env.service, env.validator, env.logger)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /chat/{chat_id}", h.GetAllMessages)

			path := "/chat/" + tc.chatIDPath
			if tc.limitQuery != "" {
				path += "?" + tc.limitQuery
			}
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)
			env.service.AssertExpectations(t)
		})
	}
}
