package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Company struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(500);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	CPVCodes    pq.StringArray `gorm:"type:text[]" json:"cpv_codes"`
	Keywords    pq.StringArray `gorm:"type:text[]" json:"keywords"`
	Countries   pq.StringArray `gorm:"type:text[]" json:"countries"`
	MinValue    *float64       `gorm:"type:decimal(15,2)" json:"min_value"`
	MaxValue    *float64       `gorm:"type:decimal(15,2)" json:"max_value"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
