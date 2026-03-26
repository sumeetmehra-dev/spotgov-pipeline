package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type Tender struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Source             string         `gorm:"type:varchar(20);not null;index" json:"source"`
	SourceID           string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"source_id"`
	Title              string         `gorm:"type:text;not null" json:"title"`
	Description        string         `gorm:"type:text" json:"description"`
	CPVCodes           pq.StringArray `gorm:"type:text[];index:idx_cpv,type:gin" json:"cpv_codes"`
	BuyerName          string         `gorm:"type:varchar(500)" json:"buyer_name"`
	BuyerCountry       string         `gorm:"type:varchar(10);index" json:"buyer_country"`
	PlaceOfPerformance string         `gorm:"type:varchar(500)" json:"place_of_performance"`
	ProcurementMethod  string         `gorm:"type:varchar(100)" json:"procurement_method"`
	EstimatedValue     *float64       `gorm:"type:decimal(15,2)" json:"estimated_value"`
	Currency           string         `gorm:"type:varchar(3);default:'EUR'" json:"currency"`
	Deadline           *time.Time     `gorm:"index" json:"deadline"`
	PublicationDate    *time.Time     `json:"publication_date"`
	NoticeURL          string         `gorm:"type:text" json:"notice_url"`
	RawData            datatypes.JSON `gorm:"type:jsonb" json:"-"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}
