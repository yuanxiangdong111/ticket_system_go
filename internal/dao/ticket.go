package dao

import (
	"ticket_system/internal/model"
	"ticket_system/pkg/database"

	"gorm.io/gorm"
)

// TicketDAO 门票数据访问对象
type TicketDAO struct {
	db *gorm.DB
}

// NewTicketDAO 创建门票DAO实例
func NewTicketDAO() *TicketDAO {
	return &TicketDAO{db: database.DB}
}

// Create 创建门票
func (d *TicketDAO) Create(ticket *model.Ticket) error {
	return d.db.Create(ticket).Error
}

// GetByID 根据ID获取门票
func (d *TicketDAO) GetByID(id uint) (*model.Ticket, error) {
	var ticket model.Ticket
	err := d.db.Preload("Category").First(&ticket, id).Error
	return &ticket, err
}

// List 获取门票列表
func (d *TicketDAO) List(categoryID, offset, limit int, status int8) ([]*model.Ticket, int64, error) {
	var tickets []*model.Ticket
	var total int64

	query := d.db.Preload("Category").Model(&model.Ticket{})
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total).Offset(offset).Limit(limit).Order("sort DESC, created_at DESC").Find(&tickets)
	return tickets, total, nil
}

// Update 更新门票
func (d *TicketDAO) Update(ticket *model.Ticket) error {
	return d.db.Save(ticket).Error
}

// UpdateStock 更新库存
func (d *TicketDAO) UpdateStock(ticketID uint, change int) error {
	return d.db.Model(&model.Ticket{}).Where("id = ?", ticketID).
		Updates(map[string]interface{}{
			"stock": gorm.Expr("stock + ?", change),
			"sold":  gorm.Expr("sold - ?", change),
		}).Error
}

// UpdateSold 更新销量
func (d *TicketDAO) UpdateSold(ticketID uint, increment int) error {
	return d.db.Model(&model.Ticket{}).Where("id = ?", ticketID).
		UpdateColumn("sold", gorm.Expr("sold + ?", increment)).Error
}

// Search 搜索门票
func (d *TicketDAO) Search(keyword string, offset, limit int) ([]*model.Ticket, int64, error) {
	var tickets []*model.Ticket
	var total int64

	query := d.db.Preload("Category").
		Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	query.Count(&total).Offset(offset).Limit(limit).Order("sold DESC, created_at DESC").Find(&tickets)
	return tickets, total, nil
}

// GetStockByID 根据ID获取库存
func (d *TicketDAO) GetStockByID(ticketID uint) (int, error) {
	var ticket model.Ticket
	err := d.db.Select("stock").Where("id = ?", ticketID).First(&ticket).Error
	if err != nil {
		return 0, err
	}
	return ticket.Stock, nil
}

// BatchUpdateStock 批量更新库存
func (d *TicketDAO) BatchUpdateStock(ticketIDs []uint, changes []int) error {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, ticketID := range ticketIDs {
		if err := tx.Model(&model.Ticket{}).Where("id = ?", ticketID).
			Updates(map[string]interface{}{
				"stock": gorm.Expr("stock + ?", changes[i]),
				"sold":  gorm.Expr("sold - ?", changes[i]),
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
