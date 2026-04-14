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

func (m *MockChatService) GetChatWithMessages(ctx context.Context, id uint, limit int) (*model.Chat, error) {
	args := m.Called(ctx, id, limit)

	var chat *model.Chat
	if args.Get(0) != nil {
		chat = args.Get(0).(*model.Chat)
	}

	return chat, args.Error(1)
}

func (m *MockChatService) DeleteChat(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChatService) CreateMessage(ctx context.Context, msg *model.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func TestChatHandler_CreateChat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()

	tests := []struct {
		name           string
		body           string
		mockSetup      func(m *MockChatService)
		expectedStatus int
	}{
		{
			name: "Success",
			body: `{"title": "Chat title"}`,
			mockSetup: func(m *MockChatService) {
				m.On(
					"CreateChat",
					mock.Anything,
					mock.MatchedBy(func(chat *model.Chat) bool {
						return chat.Title == "Chat title"
					})).Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
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
			name:           "Chat title too long (200+ chars)",
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := new(MockChatService)
			tc.mockSetup(service)

			h := handler.NewChatHandler(service, validator, logger)

			mux := http.NewServeMux()
			mux.HandleFunc("POST /chat", h.CreateChat)

			req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)
			service.AssertExpectations(t)
		})
	}
}

func TestChatHandler_DeleteChat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()

	tests := []struct {
		name           string
		chatIDPath     string
		mockSetup      func(m *MockChatService)
		expectedStatus int
	}{
		{
			name:       "Success",
			chatIDPath: "1",
			mockSetup: func(m *MockChatService) {
				m.On("DeleteChat", mock.Anything, uint(1)).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID Type",
			chatIDPath:     "abc",
			mockSetup:      func(m *MockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Not Found",
			chatIDPath: "2",
			mockSetup: func(m *MockChatService) {
				m.On("DeleteChat", mock.Anything, uint(2)).Return(model.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := new(MockChatService)
			tc.mockSetup(service)

			h := handler.NewChatHandler(service, validator, logger)

			mux := http.NewServeMux()
			mux.HandleFunc("DELETE /chat/{chat_id}", h.DeleteChat)

			req := httptest.NewRequest(http.MethodDelete, "/chat/"+tc.chatIDPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)
			service.AssertExpectations(t)
		})
	}
}
