package service_test

import (
	"time"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

func validInput() domain.WidgetInput {
	return domain.WidgetInput{
		Actor:            "ops@example.com",
		HeaderTH:         "หัวข้อ",
		StartDate:        time.Date(2025, 12, 21, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		ProductListRefID: "PL-1",
		Components: []domain.Component{
			{Type: domain.ComponentProduct, SubType: "IA", IsActive: true},
			{Type: domain.ComponentInformation, IsActive: true},
			{Type: domain.ComponentCategory, IsActive: false},
		},
	}
}
