package expand

import (
	"context"

	"github.com/ory/keto/relationtuple"
)

type (
	engineDependencies interface {
		relationtuple.ManagerProvider
	}
	Engine struct {
		d engineDependencies
	}
	EngineProvider interface {
		ExpandEngine() *Engine
	}
)

func NewEngine(d engineDependencies) *Engine {
	return &Engine{d: d}
}

func (e *Engine) BuildTree(ctx context.Context, subject relationtuple.Subject, restDepth int) (*Tree, error) {
	if restDepth <= 0 {
		return nil, nil
	}

	if us, isUserSet := subject.(*relationtuple.UserSet); isUserSet {
		subTree := &Tree{
			Type:    Union,
			Subject: subject,
		}

		rels, err := e.d.RelationTupleManager().GetRelationTuples(ctx, &relationtuple.RelationQuery{
			Relation:  us.Relation,
			ObjectID:  us.Object.ID,
			Namespace: us.Object.Namespace,
		})
		if err != nil {
			// TODO error handling
			return nil, err
		}

		if restDepth <= 1 {
			subTree.Type = Leaf
			return subTree, nil
		}

		subTree.Children = make([]*Tree, len(rels))
		for ri, r := range rels {
			subTree.Children[ri], err = e.BuildTree(ctx, r.Subject, restDepth-1)
			if err != nil {
				// TODO error handling
				return nil, err
			}
		}

		return subTree, nil
	}

	// is UserID
	return &Tree{
		Type:    Leaf,
		Subject: subject,
	}, nil
}
