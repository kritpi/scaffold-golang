package helper

import (
	"fmt"
	"net/url"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// WidgetDataSrc holds the DataSrc URL templates. config.Load checks the %s count.
type WidgetDataSrc struct {
	ProductDataSrc     string
	InformationSubType string
	CategoryDataSrc    string
}

// ResolveWidgetDataSrc fills Component.DataSrc from the templates, escaping each value.
func ResolveWidgetDataSrc(t WidgetDataSrc, widgetID string, cs []domain.Component) []domain.Component {
	id := url.PathEscape(widgetID)
	out := make([]domain.Component, len(cs))
	for i, c := range cs {
		switch c.Type {
		case domain.ComponentInformation:
			c.DataSrc = fmt.Sprintf(t.ProductDataSrc, id, url.QueryEscape(t.InformationSubType))
		case domain.ComponentProduct:
			c.DataSrc = fmt.Sprintf(t.ProductDataSrc, id, url.QueryEscape(c.SubType))
		case domain.ComponentCategory:
			c.DataSrc = fmt.Sprintf(t.CategoryDataSrc, id)
		}
		out[i] = c
	}
	return out
}
