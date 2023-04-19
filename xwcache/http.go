package xwcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"xwace/xwcache/consistenthash"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
// 每个Cache进程实例只需要一个
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self        string
	basePath    string
	mu          sync.Mutex             //把加锁放到这里，而不是在Hash里
	hashCircler *consistenthash.Hash   // 作用就是1.添加节点， 2.根据key，获得节点的机器号key
	httpClient  map[string]*httpClient // keyed by e.g. "http://10.0.0.2:8008" 多个节点
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.GetByteSlice())
}

// Set updates the pool's list of peers.
// 给这个HTTPPool去Set很多httpClient
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hashCircler = consistenthash.NewConsistentHash(defaultReplicas, nil)
	p.hashCircler.AddNode(peers...)
	p.httpClient = make(map[string]*httpClient, len(peers))
	for _, peer := range peers {
		p.httpClient[peer] = &httpClient{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
// 获得一个httpClient（向一个特定url发起请求的）
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.hashCircler.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpClient[peer], true
	}
	return nil, false
}

type httpClient struct {
	baseURL string
}

func (h *httpClient) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	log.Printf("[http.get %s] bytes=%v", u, bytes)
	return bytes, nil
}
