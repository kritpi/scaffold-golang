// Package helper holds the functions that two or more use cases share.
package helper

import (
	"context"
	"fmt"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
	"github.com/kritpi/scaffold-golang/internal/core/port"
)

// CheckWidgetInput validates in and checks that its product list exists.
func CheckWidgetInput(ctx context.Context, catalog port.CatalogAdapter, in domain.WidgetInput) error {
	if err := in.Validate(); err != nil {
		return err
	}
	if in.ProductListRefID == "" {
		return nil
	}
	ok, err := catalog.ProductListExists(ctx, in.ProductListRefID)
	if err != nil {
		return fmt.Errorf("check product list %s: %w", in.ProductListRefID, err)
	}
	if !ok {
		return fmt.Errorf("%w: product_list_ref_id %q does not exist", domain.ErrInvalidInput, in.ProductListRefID)
	}
	return nil
}
