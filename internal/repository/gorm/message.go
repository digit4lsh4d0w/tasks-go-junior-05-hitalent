package gorm

import (
	"task-5/internal/log"
	"task-5/internal/model"
	"task-5/internal/repository"

	"gorm.io/gorm"
)

type msgRepo struct {
	db     *gorm.DB
	logger log.Logger
}

func NewMessageRepository(db *gorm.DB, logger log.Logger) repository.MessageRepository {
	logger.Debug("Creating new message repository", "type", "gorm")
	return &msgRepo{db, logger}
}

func (r *msgRepo) FindAll() ([]model.Message, error) {
	var msgs []model.Message
	result := r.db.Find(&msgs)
	return msgs, result.Error
}

func (r *msgRepo) FindByID(id uint) (*model.Message, error) {
	var msg model.Message
	result := r.db.Find(&msg, id)
	return &msg, result.Error
}

func (r *msgRepo) Create(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *msgRepo) Update(msg *model.Message) error {
	return r.db.Save(msg).Error
}

func (r *msgRepo) Delete(id uint) error {
	return r.db.Delete(&model.Message{}, id).Error
}
