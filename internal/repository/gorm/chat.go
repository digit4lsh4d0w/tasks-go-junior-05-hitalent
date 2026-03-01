package gorm

import (
	"task-5/internal/log"
	"task-5/internal/model"
	"task-5/internal/repository"

	"gorm.io/gorm"
)

type chatRepo struct {
	db  *gorm.DB
	log log.Logger
}

func NewChatRepository(db *gorm.DB, log log.Logger) repository.ChatRepository {
	log.Debug("Creating new chat repository", "type", "gorm")
	return &chatRepo{db, log}
}

func (r *chatRepo) FindAll() ([]model.Chat, error) {
	var chats []model.Chat
	result := r.db.Find(&chats)
	return chats, result.Error
}

func (r *chatRepo) FindByID(id uint) (*model.Chat, error) {
	var chat model.Chat
	result := r.db.Preload("Messages").First(&chat, id)
	return &chat, result.Error
}

func (r *chatRepo) Create(chat *model.Chat) error {
	return r.db.Create(chat).Error
}

func (r *chatRepo) Update(chat *model.Chat) error {
	return r.db.Save(chat).Error
}

func (r *chatRepo) Delete(id uint) error {
	return r.db.Delete(&model.Chat{}, id).Error
}
