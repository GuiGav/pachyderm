package server

import (
	etcd "github.com/coreos/etcd/clientv3"
	"github.com/jmoiron/sqlx"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Env is the dependencies needed to run the PFS API server
type Env struct {
	ObjectClient obj.Client
	DB           *sqlx.DB
	EtcdPrefix   string
	EtcdClient   *etcd.Client
	TxnEnv       *txnenv.TransactionEnv
	Listener     col.PostgresListener

	AuthServer authserver.APIServer
	// TODO: a reasonable repo metadata solution would let us get rid of this circular dependency
	// permissions might also work.
	PPSServer ppsserver.APIServer

	BackgroundContext context.Context
	Logger            *logrus.Logger
	StorageConfig     serviceenv.StorageConfiguration
}

func EnvFromServiceEnv(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv) (*Env, error) {
	// Setup etcd, object storage, and database clients.
	objClient, err := obj.NewClient(env.Config().StorageBackend, env.Config().StorageRoot)
	if err != nil {
		return nil, err
	}

	return &Env{
		ObjectClient: objClient,
		DB:           env.GetDBClient(),
		TxnEnv:       txnEnv,
		Listener:     env.GetPostgresListener(),
		EtcdPrefix:   env.Config().PFSEtcdPrefix,
		EtcdClient:   env.GetEtcdClient(),

		AuthServer: env.AuthServer(),
		PPSServer:  env.PpsServer(),

		BackgroundContext: env.Context(),
		StorageConfig:     env.Config().StorageConfiguration,
		Logger:            logrus.StandardLogger(),
	}, nil
}
