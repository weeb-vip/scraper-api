package thetvdblink

import (
	"time"
)

type TheTVDBLink struct {
	ID            string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AnimeID       string    `gorm:"column:anime_id;type:uuid;not null" json:"anime_id"`
	TheTVDBLinkID string    `gorm:"column:thetvdb_id;not null" json:"thetvdb_link_id"`
	Season        int       `gorm:"column:season_number;not null" json:"season"`
	Name          string    `gorm:"column:name;not null" json:"name"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// set table name
func (TheTVDBLink) TableName() string {
	return "thetvdb_link"
}
