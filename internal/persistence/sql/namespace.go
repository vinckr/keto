package sql

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/gobuffalo/pop/v5"

	"github.com/ory/keto/internal/namespace"
)

type (
	namespaceRow struct {
		ID      int    `db:"id"`
		Name    string `db:"name"`
		Version int    `db:"schema_version"`
	}
)

const (
	namespaceCreateStatement = `
CREATE TABLE %[1]s
(
    shard_id    varchar(64),
    object      varchar(64),
    relation    varchar(64),
    subject     varchar(256), /* can be <namespace:object#relation> or <user_id> */
    commit_time timestamp,

	PRIMARY KEY (shard_id, object, relation, subject, commit_time)
);

CREATE INDEX %[1]s_object_idx ON %[1]s (object);

CREATE INDEX %[1]s_user_set_idx ON %[1]s (object, relation);
`

	mostRecentSchemaVersion = 1
)

func tableFromNamespace(n *namespace.Namespace) string {
	return fmt.Sprintf("keto_%0.10d_relation_tuples", n.ID)
}

func createStmt(n *namespace.Namespace) string {
	return fmt.Sprintf(namespaceCreateStatement, tableFromNamespace(n))
}

func (p *Persister) MigrateNamespaceUp(ctx context.Context, n *namespace.Namespace) error {
	return p.transaction(ctx, func(ctx context.Context, c *pop.Connection) error {
		// TODO this is only creating new namespaces and not applying migrations

		if err := c.RawQuery("INSERT INTO keto_namespace (id, name, schema_version) VALUES (?, ?, ?)", n.ID, n.Name, mostRecentSchemaVersion).Exec(); err != nil {
			return errors.WithStack(err)
		}

		return errors.WithStack(
			c.RawQuery(createStmt(n)).Exec())
	})
}

func (p *Persister) NamespaceFromName(ctx context.Context, name string) (*namespace.Namespace, error) {
	var n namespace.Namespace

	return &n, errors.WithStack(
		p.connection(ctx).Where("name = ?", name).First(&n))
}

func (p *Persister) NamespaceStatus(ctx context.Context, name string) (*namespace.Status, error) {
	var n namespaceRow
	if err := p.connection(ctx).Where("name = ?", name).First(&n); err != nil {
		return nil, err
	}

	return &namespace.Status{
		CurrentVersion: n.Version,
		NextVersion:    mostRecentSchemaVersion,
	}, nil
}

func (n *namespaceRow) TableName() string {
	return "keto_namespace"
}
