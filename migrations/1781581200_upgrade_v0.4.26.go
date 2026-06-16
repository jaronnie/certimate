package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.26")
		tracer.Printf("go ...")

		if collection, err := app.FindCollectionByNameOrId("monitor"); err == nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
			tracer.Printf("collection 'monitor' deleted")
		}

		if err := createMonitorCertificateCollection(app); err != nil {
			return err
		}
		if err := createMonitorDomainCollection(app); err != nil {
			return err
		}

		tracer.Printf("monitor collections updated")
		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
