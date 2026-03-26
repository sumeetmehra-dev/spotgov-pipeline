package model

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_company_tender" json:"company_id"`
	TenderID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_company_tender" json:"tender_id"`
	Score       float64   `gorm:"type:decimal(5,4);not null" json:"score"`
	VectorScore float64   `gorm:"type:decimal(5,4)" json:"vector_score"`
	BM25Score   float64   `gorm:"type:decimal(5,4)" json:"bm25_score"`
	CPVOverlap  float64   `gorm:"type:decimal(5,4)" json:"cpv_overlap"`
	MatchedAt   time.Time `gorm:"default:now()" json:"matched_at"`

	Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"-"`
	Tender  Tender  `gorm:"foreignKey:TenderID;constraint:OnDelete:CASCADE" json:"tender,omitempty"`
}
