package driver

import (
	"context"

	"github.com/gobuffalo/pop/v5"
	"github.com/ory/herodot"
	"github.com/ory/x/logrusx"

	"github.com/ory/keto/namespace"
	"github.com/ory/keto/persistence/sql"

	"github.com/ory/keto/persistence"

	"github.com/ory/keto/expand"

	"github.com/ory/keto/check"

	"github.com/ory/keto/relationtuple"
	"github.com/ory/keto/x"
)

var _ relationtuple.ManagerProvider = &RegistryDefault{}
var _ x.WriterProvider = &RegistryDefault{}
var _ x.LoggerProvider = &RegistryDefault{}

type RegistryDefault struct {
	p    *sql.Persister
	l    *logrusx.Logger
	w    herodot.Writer
	ce   *check.Engine
	ee   *expand.Engine
	conn *pop.Connection
}

func (r *RegistryDefault) Logger() *logrusx.Logger {
	if r.l == nil {
		r.l = logrusx.New("keto", "dev")
	}
	return r.l
}

func (r *RegistryDefault) Writer() herodot.Writer {
	if r.w == nil {
		r.w = herodot.NewJSONWriter(r.Logger())
	}
	return r.w
}

func (r *RegistryDefault) RelationTupleManager() relationtuple.Manager {
	return r.p
}

func (r *RegistryDefault) NamespaceManagerProvider() namespace.Manager {
	return r.p
}

func (r *RegistryDefault) PermissionEngine() *check.Engine {
	if r.ce == nil {
		r.ce = check.NewEngine(r)
	}
	return r.ce
}

func (r *RegistryDefault) ExpandEngine() *expand.Engine {
	if r.ee == nil {
		r.ee = expand.NewEngine(r)
	}
	return r.ee
}

func (r *RegistryDefault) Persister() (persistence.Persister, error) {
	if r.p == nil {
		var err error
		r.p, err = sql.NewPersister(r.conn)
		if err != nil {
			return nil, err
		}
	}
	return r.p, nil
}

func (r *RegistryDefault) Migrator() (persistence.Migrator, error) {
	if r.p == nil {
		if _, err := r.Persister(); err != nil {
			return nil, err
		}
	}
	return r.p, nil
}

func (r *RegistryDefault) Init() error {
	c, err := pop.NewConnection(&pop.ConnectionDetails{
		URL: "sqlite://:memory:?_fk=true",
	})
	r.conn = c
	m, err := r.Migrator()
	if err != nil {
		return err
	}
	return m.MigrateUp(context.Background())
}
