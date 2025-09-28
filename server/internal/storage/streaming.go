package storage

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net/http"
    "sync"
)

// car стример для больших файлов
type CARStreamer struct {
    ipfsEndpoint string
    httpClient   *http.Client
    chunkSize    int
}

func NewCARStreamer(endpoint string, chunkSize int) *CARStreamer {
    if chunkSize <= 0 {
        chunkSize = 1024 * 1024 // 1mb по умолчанию
    }
    
    return &CARStreamer{
        ipfsEndpoint: endpoint,
        httpClient: &http.Client{
            Timeout: 0, // без таймаута для стриминга
        },
        chunkSize: chunkSize,
    }
}

// стримить car из ipfs
func (s *CARStreamer) StreamCAR(ctx context.Context, cid string) (<-chan []byte, <-chan error) {
    chunkChan := make(chan []byte, 2) // буфер для backpressure
    errChan := make(chan error, 1)
    
    go func() {
        defer close(chunkChan)
        defer close(errChan)
        
        // запрос к ipfs
        url := fmt.Sprintf("%s/api/v0/dag/export?arg=%s", s.ipfsEndpoint, cid)
        req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
        if err != nil {
            errChan <- err
            return
        }
        
        resp, err := s.httpClient.Do(req)
        if err != nil {
            errChan <- err
            return
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            errChan <- fmt.Errorf("ipfs error: %s", resp.Status)
            return
        }
        
        // читаем чанками
        buffer := make([]byte, s.chunkSize)
        for {
            select {
            case <-ctx.Done():
                errChan <- ctx.Err()
                return
            default:
            }
            
            n, err := resp.Body.Read(buffer)
            if n > 0 {
                chunk := make([]byte, n)
                copy(chunk, buffer[:n])
                
                select {
                case chunkChan <- chunk:
                    // отправили чанк
                case <-ctx.Done():
                    errChan <- ctx.Err()
                    return
                }
            }
            
            if err != nil {
                if err != io.EOF {
                    errChan <- err
                }
                return
            }
        }
    }()
    
    return chunkChan, errChan
}

// загрузить car стримом
func (s *CARStreamer) UploadCAR(ctx context.Context, chunks <-chan []byte) (string, error) {
    // создаем pipe для стриминга
    pr, pw := io.Pipe()
    
    var wg sync.WaitGroup
    var uploadErr error
    var cid string
    
    // горутина для записи чанков в pipe
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer pw.Close()
        
        for chunk := range chunks {
            select {
            case <-ctx.Done():
                pw.CloseWithError(ctx.Err())
                return
            default:
            }
            
            if _, err := pw.Write(chunk); err != nil {
                pw.CloseWithError(err)
                return
            }
        }
    }()
    
    // горутина для загрузки в ipfs
    wg.Add(1)
    go func() {
        defer wg.Done()
        
        url := fmt.Sprintf("%s/api/v0/dag/import", s.ipfsEndpoint)
        req, err := http.NewRequestWithContext(ctx, "POST", url, pr)
        if err != nil {
            uploadErr = err
            return
        }
        
        req.Header.Set("Content-Type", "application/vnd.ipld.car")
        
        resp, err := s.httpClient.Do(req)
        if err != nil {
            uploadErr = err
            return
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            uploadErr = fmt.Errorf("ipfs error: %s", resp.Status)
            return
        }
        
        // парсим cid из ответа
        // тут должен быть json парсинг
        // пока заглушка
        cid = "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
    }()
    
    wg.Wait()
    
    if uploadErr != nil {
        return "", uploadErr
    }
    
    if cid == "" {
        return "", errors.New("no cid returned")
    }
    
    return cid, nil
}

// верифицировать car стримом
func (s *CARStreamer) VerifyCAR(ctx context.Context, chunks <-chan []byte, expectedHash []byte) error {
    // тут должна быть blake3 проверка по чанкам
    // пока заглушка
    
    for range chunks {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
    }
    
    return nil
}

// car reader для чтения с лимитами
type CARReader struct {
    source    io.ReadCloser
    bytesRead int64
    maxBytes  int64
}

func NewCARReader(source io.ReadCloser, maxBytes int64) *CARReader {
    return &CARReader{
        source:   source,
        maxBytes: maxBytes,
    }
}

func (r *CARReader) Read(p []byte) (n int, err error) {
    if r.bytesRead >= r.maxBytes {
        return 0, errors.New("size limit exceeded")
    }
    
    remaining := r.maxBytes - r.bytesRead
    if int64(len(p)) > remaining {
        p = p[:remaining]
    }
    
    n, err = r.source.Read(p)
    r.bytesRead += int64(n)
    
    return n, err
}

func (r *CARReader) Close() error {
    return r.source.Close()
}
