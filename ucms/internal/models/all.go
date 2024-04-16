package models

import (
	"uvoo.io/ucms/internal/database"
)

func Migrate() {
	if err := database.DBCon.AutoMigrate(
		&Page{},
		&CountryCodeRule{},
		&FWRule{},
		&User{},
		&JwtCustomClaims{},
	); err != nil {
		panic(err)
		// e.Logger.Fatal(err)
	}

	if err := database.DBCon.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_direction__priority ON fw_rules (direction, priority)").Error; err != nil {
		panic("failed to create unique index")
	}
	if err := database.DBCon.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_direction__action__src_ip_net ON fw_rules (direction, action, src_ip_net)").Error; err != nil {
		panic("failed to create unique index")
	}
}
