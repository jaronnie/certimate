package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.26")
		tracer.Printf("go ...")

		if err := createMonitorCertificateCollection(app); err != nil {
			return err
		}
		if err := createMonitorDomainCollection(app); err != nil {
			return err
		}

		tracer.Printf("monitor collections created")
		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		if collection, err := app.FindCollectionByNameOrId("monitor_certificate"); err == nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
		}
		if collection, err := app.FindCollectionByNameOrId("monitor_domain"); err == nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
		}
		return nil
	})
}

func createMonitorCertificateCollection(app core.App) error {
	if _, err := app.FindCollectionByNameOrId("monitor_certificate"); err == nil {
		return nil
	}

	collection := core.NewBaseCollection("monitor_certificate")
	collection.Id = "monitorcert0000"
	collection.Fields.Add(&core.TextField{Name: "monitorDomain", Required: true})
	collection.Fields.Add(&core.TextField{Name: "certificateSubjectAltNames", Required: true})
	collection.Fields.Add(&core.TextField{Name: "certificateSerialNumber"})
	collection.Fields.Add(&core.TextField{Name: "certificateIssuerName"})
	collection.Fields.Add(&core.DateField{Name: "certificateNotBefore"})
	collection.Fields.Add(&core.DateField{Name: "certificateNotAfter", Required: true})
	collection.Fields.Add(&core.RelationField{Name: "workflowRef", CollectionId: "tovyif5ax6j62ur", MaxSelect: 1})
	collection.Fields.Add(&core.RelationField{Name: "workflowRunRef", CollectionId: "qjp8lygssgwyqyz", MaxSelect: 1})
	collection.Fields.Add(&core.TextField{Name: "workflowNodeId"})
	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})

	if err := app.Save(collection); err != nil {
		return err
	}
	if _, err := app.DB().NewQuery("CREATE UNIQUE INDEX idx_monitor_certificate_monitor_domain ON monitor_certificate (monitorDomain)").Execute(); err != nil {
		return err
	}
	if _, err := app.DB().NewQuery("CREATE INDEX idx_monitor_certificate_subject_alt_names ON monitor_certificate (certificateSubjectAltNames)").Execute(); err != nil {
		return err
	}
	return nil
}

func createMonitorDomainCollection(app core.App) error {
	if _, err := app.FindCollectionByNameOrId("monitor_domain"); err == nil {
		return nil
	}

	collection := core.NewBaseCollection("monitor_domain")
	collection.Id = "monitordomain00"
	collection.Fields.Add(&core.TextField{Name: "monitorDomain", Required: true})
	collection.Fields.Add(&core.TextField{Name: "domainName", Required: true})
	collection.Fields.Add(&core.DateField{Name: "domainNotAfter", Required: true})
	collection.Fields.Add(&core.RelationField{Name: "workflowRef", CollectionId: "tovyif5ax6j62ur", MaxSelect: 1})
	collection.Fields.Add(&core.RelationField{Name: "workflowRunRef", CollectionId: "qjp8lygssgwyqyz", MaxSelect: 1})
	collection.Fields.Add(&core.TextField{Name: "workflowNodeId"})
	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})

	if err := app.Save(collection); err != nil {
		return err
	}
	if _, err := app.DB().NewQuery("CREATE UNIQUE INDEX idx_monitor_domain_domain_name ON monitor_domain (domainName)").Execute(); err != nil {
		return err
	}
	if _, err := app.DB().NewQuery("CREATE INDEX idx_monitor_domain_monitor_domain ON monitor_domain (monitorDomain)").Execute(); err != nil {
		return err
	}
	return nil
}
