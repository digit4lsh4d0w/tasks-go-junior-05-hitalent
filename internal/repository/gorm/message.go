package gorm

import (
	"task-5/internal/log"
	"task-5/internal/model"
	"task-5/internal/repository"

	"gorm.io/gorm"
)

type messageRepository struct {
	db     *gorm.DB
	logger log.Logger
}

func NewMessageRepository(db *gorm.DB, logger log.Logger) repository.MessageRepository {
	logger.Debug("Creating new message repository", "type", "gorm")
	return &messageRepository{db, logger}
}

func (r *messageRepository) FindAll() ([]model.Message, error) {
	var msgs []model.Message
	result := r.db.Find(&msgs)
	return msgs, result.Error
}

func (r *messageRepository) FindByID(id uint) (*model.Message, error) {
	var msg model.Message
	result := r.db.Find(&msg, id)
	return &msg, result.Error
}

func (r *messageRepository) Create(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *messageRepository) Update(msg *model.Message) error {
	return r.db.Save(msg).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&model.Message{}, id).Error
}
