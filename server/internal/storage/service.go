package storage

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"dev.c0rex64.heroin/internal/ipfs"
)

type Service struct {
	ipfs     *ipfs.Client
	pin      bool
	replicas int
	db       *sql.DB
}

func New(ipfsClient *ipfs.Client, pin bool, replicas int) *Service {
	return &Service{ipfs: ipfsClient, pin: pin, replicas: replicas}
}

func NewWithDB(ipfsClient *ipfs.Client, pin bool, replicas int, db *sql.DB) *Service {
	return &Service{ipfs: ipfsClient, pin: pin, replicas: replicas, db: db}
}

func (s *Service) PutCAR(ctx context.Context, name, mime string, size int64, carChunks [][]byte, totalBlake3 []byte) (string, string, error) {
	buf := bytes.Join(carChunks, nil)
	if len(totalBlake3) > 0 {
		calc := blake3Sum(buf)
		if !bytes.Equal(calc, totalBlake3) {
			return "", "", fmt.Errorf("blake3 mismatch")
		}
	}
	cid, err := s.ipfs.AddCAR(ctx, buf)
	if err != nil { return "", "", err }
	if s.pin {
		if err := s.ipfs.PinAdd(ctx, cid); err != nil { return "", "", err }
	}
	if s.db != nil {
		fileID := generateFileID()
		_, err := s.db.ExecContext(ctx, `INSERT INTO files (id, user_id, cid, name, mime, size_bytes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			fileID, "system", cid, name, mime, size, time.Now().Unix())
		if err != nil { return "", "", fmt.Errorf("store file metadata: %w", err) }
		return fileID, cid, nil
	}
	return cid, cid, nil
}

func (s *Service) GetCAR(ctx context.Context, cid string) ([][]byte, error) {
	r, err := s.ipfs.ExportCAR(ctx, cid)
	if err != nil { return nil, err }
	defer r.Close()
	const chunkSize = 64 * 1024
	var chunks [][]byte
	buf := make([]byte, chunkSize)
	for {
		n, readErr := io.ReadFull(r, buf)
		if n > 0 {
			copyBuf := make([]byte, n)
			copy(copyBuf, buf[:n])
			chunks = append(chunks, copyBuf)
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}
	return chunks, nil
}

func generateFileID() string {
	now := time.Now().UnixMilli()
	b := make([]byte, 16)
	for i := 0; i < 6; i++ { b[i] = byte(now >> (8 * (5 - i))) }
	b[6] = 0x70 | 0x0F&b[6]
	b[8] = 0x80 | 0x3F&b[8]
	for i := 9; i < 16; i++ { b[i] = byte(time.Now().UnixNano() >> (8 * (i % 8))) }
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func blake3Sum(b []byte) []byte {
	// simple placeholder, TODO: replace with real blake3 hash computation
	// use sha256 as temp to avoid pulling blake3 dep right now
	// replace with a proper blake3 library in the next step
	return pseudoHash(b)
}

func pseudoHash(b []byte) []byte {
	var h uint64
	for _, v := range b { h = h*131 + uint64(v) }
	out := make([]byte, 32)
	for i := 0; i < 32; i++ { out[i] = byte(h >> uint((i%8)*8)) }
	return out
}
