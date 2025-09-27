package grpcapi

import (
	"dev.c0rex64.heroin/internal/ipfs"
	"dev.c0rex64.heroin/internal/messaging"
	"dev.c0rex64.heroin/internal/storage"
	"database/sql"
)

func (s *Server) WireStorageAndMessaging(ipfsEndpoint string, pin bool, replicas int, db *sql.DB, kp KeyProvider) {
	ic := ipfs.New(ipfsEndpoint)
	st := storage.NewWithDB(ic, pin, replicas, db)
	q := messaging.NewQueue(db)
	ms := messaging.NewService(q, kp)
	s.StorageSvc = st
	s.MessagingSvc = ms
}
