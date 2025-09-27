package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"time"
)

type Client struct {
	endpoint string
	http     *http.Client
}

type PinAddResponse struct {
	Pins []string `json:"Pins"`
}

type PinAddRequest struct {
	CID string
}

func New(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) PinAdd(ctx context.Context, cid string) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "/api/v0/pin/add")
	q := u.Query()
	q.Set("arg", cid)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pin add failed: %s", string(b))
	}
	return nil
}

func (c *Client) AddCAR(ctx context.Context, car []byte) (string, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "/api/v0/dag/import")
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("file", "data.car")
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(car); err != nil {
		return "", err
	}
	w.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("dag import failed: %s", string(b))
	}
	var v struct {
		Root struct {
			Cid string `json:"/"`
		} `json:"Root"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&v); err != nil {
		return "", err
	}
	return v.Root.Cid, nil
}

func (c *Client) ExportCAR(ctx context.Context, cid string) (io.ReadCloser, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/v0/dag/export")
	q := u.Query()
	q.Set("arg", cid)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("dag export failed: %s", string(b))
	}
	return resp.Body, nil
}
