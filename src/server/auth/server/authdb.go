package server

import (
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
)

const (
	authConfigCollectionName   = "auth_config"
	roleBindingsCollectionName = "role_bindings"
	membersCollectionName      = "members"
	groupsCollectionName       = "groups"
)

var authConfigIndexes = []*col.Index{}

func authConfigCollection(db *sqlx.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		authConfigCollectionName,
		db,
		listener,
		&auth.OIDCConfig{},
		nil,
	)
}

var roleBindingsIndexes = []*col.Index{}

func roleBindingsCollection(db *sqlx.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		roleBindingsCollectionName,
		db,
		listener,
		&auth.RoleBinding{},
		roleBindingsIndexes,
	)
}

var membersIndexes = []*col.Index{}

func membersCollection(db *sqlx.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		membersCollectionName,
		db,
		listener,
		&auth.Groups{},
		membersIndexes,
	)
}

var groupsIndexes = []*col.Index{}

func groupsCollection(db *sqlx.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		groupsCollectionName,
		db,
		listener,
		&auth.Users{},
		groupsIndexes,
	)
}

// AllCollections returns a list of all the PPS API collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(authConfigCollectionName, nil, nil, nil, authConfigIndexes),
		col.NewPostgresCollection(roleBindingsCollectionName, nil, nil, nil, roleBindingsIndexes),
		col.NewPostgresCollection(membersCollectionName, nil, nil, nil, membersIndexes),
		col.NewPostgresCollection(groupsCollectionName, nil, nil, nil, groupsIndexes),
	}
}