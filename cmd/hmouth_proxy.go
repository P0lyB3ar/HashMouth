package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hashmouth/network"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HMouthProxy is a local proxy that resolves .hmouth domains
type HMouthProxy struct {
	dht           *network.DHT
	node          *network.P2PNode
	relayNet      *network.RelayNetwork
	sharedKey     []byte
	nodeID        string
	domains       map[string]*HMouthDomain // domain -> info
	hostedSites   map[string]*HostedSite   // our hosted sites
	proxyPort     string
	mu            sync.RWMutex
}

// HMouthDomain represents a .hmouth domain
type HMouthDomain struct {
	Domain    string    `json:"domain"`    // e.g., "mysite.hmouth"
	NodeID    string    `json:"nodeId"`    // Hosting node
	Addr      string    `json:"addr"`      // Node address
	PublicKey string    `json:"publicKey"` // For verification
	LastSeen  time.Time `json:"lastSeen"`
}

// HostedSite represents a site we're hosting
type HostedSite struct {
	Domain      string
	ContentPath string
	BackendURL  string // For proxying to backend (e.g., "http://localhost:3000")
	Handler     http.Handler
	IsBackend   bool
}

func generateHMouthDomain() string {
	b := make([]byte, 16)
	cryptorand.Read(b)
	return hex.EncodeToString(b) + ".hmouth"
}

func NewHMouthProxy(dhtPort, p2pPort int, proxyPort string) (*HMouthProxy, error) {
	nodeID := generateNodeID()

	// Start DHT
	dht, err := network.NewDHT(dhtPort)
	if err != nil {
		return nil, fmt.Errorf("failed to start DHT: %v", err)
	}

	// Start P2P
	p2pAddr := fmt.Sprintf(":%d", p2pPort)
	node := network.NewNode(nodeID, p2pAddr)
	if err := node.Listen(); err != nil {
		return nil, fmt.Errorf("failed to start P2P: %v", err)
	}

	// Start relay network
	relayNet := network.NewRelayNetwork()
	relayNet.RegisterRelayNode(nodeID, p2pAddr)
	relayNet.StartCleanupRoutine()

	sharedKey := []byte("12345678901234567890123456789012")

	proxy := &HMouthProxy{
		dht:         dht,
		node:        node,
		relayNet:    relayNet,
		sharedKey:   sharedKey,
		nodeID:      nodeID,
		domains:     make(map[string]*HMouthDomain),
		hostedSites: make(map[string]*HostedSite),
		proxyPort:   proxyPort,
	}

	// Bootstrap DHT
	log.Printf("🌐 Connecting to DHT network...")
	if err := dht.Bootstrap(); err != nil {
		log.Printf("⚠️  DHT bootstrap warning: %v", err)
	}

	// Start domain discovery
	go proxy.discoverDomains()
	go proxy.announceDomains()

	return proxy, nil
}

func generateNodeID() string {
	b := make([]byte, 20)
	cryptorand.Read(b)
	return hex.EncodeToString(b)
}

// HostSite hosts a new .hmouth site (static files)
func (hp *HMouthProxy) HostSite(contentPath string, customDomain string) (string, error) {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	domain := customDomain
	if domain == "" {
		domain = generateHMouthDomain()
	} else if !strings.HasSuffix(domain, ".hmouth") {
		domain = domain + ".hmouth"
	}

	// Create file server for content
	handler := http.FileServer(http.Dir(contentPath))

	site := &HostedSite{
		Domain:      domain,
		ContentPath: contentPath,
		Handler:     handler,
		IsBackend:   false,
	}

	hp.hostedSites[domain] = site

	// Register domain in DHT
	domainInfo := &HMouthDomain{
		Domain:    domain,
		NodeID:    hp.nodeID,
		Addr:      hp.node.Addr,
		PublicKey: hp.nodeID[:32], // Simplified
		LastSeen:  time.Now(),
	}

	hp.domains[domain] = domainInfo

	log.Printf("🌐 Hosting static site: %s", domain)
	log.Printf("📁 Content path: %s", contentPath)
	log.Printf("🔗 Access via: http://%s (through proxy)", domain)

	return domain, nil
}

// HostBackend hosts a backend application (proxies to local server)
func (hp *HMouthProxy) HostBackend(backendURL string, customDomain string) (string, error) {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	domain := customDomain
	if domain == "" {
		domain = generateHMouthDomain()
	} else if !strings.HasSuffix(domain, ".hmouth") {
		domain = domain + ".hmouth"
	}

	// Create reverse proxy handler
	handler := hp.createReverseProxy(backendURL)

	site := &HostedSite{
		Domain:     domain,
		BackendURL: backendURL,
		Handler:    handler,
		IsBackend:  true,
	}

	hp.hostedSites[domain] = site

	// Register domain in DHT
	domainInfo := &HMouthDomain{
		Domain:    domain,
		NodeID:    hp.nodeID,
		Addr:      hp.node.Addr,
		PublicKey: hp.nodeID[:32], // Simplified
		LastSeen:  time.Now(),
	}

	hp.domains[domain] = domainInfo

	log.Printf("🌐 Hosting backend: %s", domain)
	log.Printf("🔗 Backend URL: %s", backendURL)
	log.Printf("🔗 Access via: http://%s (through proxy)", domain)

	return domain, nil
}

// createReverseProxy creates a reverse proxy to backend
func (hp *HMouthProxy) createReverseProxy(backendURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create new request to backend
		backendReq, err := http.NewRequest(r.Method, backendURL+r.URL.Path, r.Body)
		if err != nil {
			http.Error(w, "Failed to create backend request", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for key, values := range r.Header {
			for _, value := range values {
				backendReq.Header.Add(key, value)
			}
		}

		// Copy query parameters
		backendReq.URL.RawQuery = r.URL.RawQuery

		// Send request to backend
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(backendReq)
		if err != nil {
			http.Error(w, "Backend unavailable: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Copy status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)
	})
}

// discoverDomains watches for new .hmouth domains on the network
func (hp *HMouthProxy) discoverDomains() {
	peerCh := hp.dht.GetPeerChannel()

	for peer := range peerCh {
		// Connect to peer
		peerAddr := fmt.Sprintf("%s:%d", peer.Addr, peer.Port)
		hp.node.ConnectPeer(peer.ID, peerAddr)
		hp.relayNet.RegisterRelayNode(peer.ID, peerAddr)

		// Request their hosted domains
		go hp.requestDomains(peer.ID)
	}
}

func (hp *HMouthProxy) requestDomains(peerID string) {
	// In a real implementation, this would query the peer for their domains
	// For now, domains are discovered through DHT announcements
}

// announceDomains announces our hosted domains to the network
func (hp *HMouthProxy) announceDomains() {
	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hp.mu.RLock()
			domainCount := len(hp.hostedSites)
			hp.mu.RUnlock()

			if domainCount > 0 {
				hp.dht.Announce()
				log.Printf("📢 Announced %d .hmouth domains", domainCount)
			}
		}
	}
}

// ResolveDomain resolves a .hmouth domain to content
func (hp *HMouthProxy) ResolveDomain(domain string) (http.Handler, error) {
	hp.mu.RLock()
	defer hp.mu.RUnlock()

	// Check if we're hosting it
	if site, exists := hp.hostedSites[domain]; exists {
		return site.Handler, nil
	}

	// Check if we know about it
	if domainInfo, exists := hp.domains[domain]; exists {
		// Fetch from remote node
		return hp.createRemoteHandler(domainInfo), nil
	}

	return nil, fmt.Errorf("domain not found: %s", domain)
}

// createRemoteHandler creates a handler that fetches content from remote node
func (hp *HMouthProxy) createRemoteHandler(domainInfo *HMouthDomain) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fetch content from remote node through relay network
		content, err := hp.fetchRemoteContent(domainInfo, r.URL.Path)
		if err != nil {
			http.Error(w, "Failed to fetch content: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Serve the content
		w.Header().Set("Content-Type", detectContentType(r.URL.Path))
		w.Write(content)
	})
}

func (hp *HMouthProxy) fetchRemoteContent(domainInfo *HMouthDomain, path string) ([]byte, error) {
	// In a real implementation, this would:
	// 1. Build a relay path to the hosting node
	// 2. Send an encrypted request for the content
	// 3. Receive and decrypt the response
	// For now, return a placeholder
	return []byte(fmt.Sprintf("<html><body><h1>%s</h1><p>Content from remote node (path: %s)</p></body></html>",
		domainInfo.Domain, path)), nil
}

func detectContentType(path string) string {
	if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") {
		return "text/html"
	} else if strings.HasSuffix(path, ".css") {
		return "text/css"
	} else if strings.HasSuffix(path, ".js") {
		return "application/javascript"
	} else if strings.HasSuffix(path, ".json") {
		return "application/json"
	} else if strings.HasSuffix(path, ".png") {
		return "image/png"
	} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		return "image/jpeg"
	}
	return "text/html"
}

// StartProxy starts the HTTP proxy server
func (hp *HMouthProxy) StartProxy() error {
	mux := http.NewServeMux()

	// Proxy handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if host == "" {
			host = r.Header.Get("Host")
		}

		// Remove port if present
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Check if it's a .hmouth domain
		if strings.HasSuffix(host, ".hmouth") {
			handler, err := hp.ResolveDomain(host)
			if err != nil {
				http.Error(w, "Domain not found: "+host, http.StatusNotFound)
				return
			}
			handler.ServeHTTP(w, r)
			return
		}

		// Control panel
		hp.serveControlPanel(w, r)
	})

	// API endpoints
	mux.HandleFunc("/api/host", hp.handleHostSite)
	mux.HandleFunc("/api/host-backend", hp.handleHostBackend)
	mux.HandleFunc("/api/domains", hp.handleListDomains)
	mux.HandleFunc("/api/stats", hp.handleStats)

	log.Printf("🚀 HMouth Proxy started on http://localhost%s", hp.proxyPort)
	log.Printf("📋 Control panel: http://localhost%s", hp.proxyPort)
	log.Printf("🌐 Configure your browser to use this proxy")
	log.Printf("")
	log.Printf("Firefox Proxy Settings:")
	log.Printf("  1. Open Settings → Network Settings")
	log.Printf("  2. Manual proxy configuration")
	log.Printf("  3. HTTP Proxy: localhost, Port: %s", strings.TrimPrefix(hp.proxyPort, ":"))
	log.Printf("  4. Check 'Also use this proxy for HTTPS'")
	log.Printf("")

	return http.ListenAndServe(hp.proxyPort, mux)
}

func (hp *HMouthProxy) serveControlPanel(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>HMouth Proxy Control Panel</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            padding: 30px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
        }
        h1 {
            color: #667eea;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
        }
        .section {
            margin-bottom: 30px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 10px;
        }
        .section h2 {
            color: #333;
            margin-bottom: 15px;
        }
        input, button {
            padding: 12px;
            border-radius: 8px;
            border: 1px solid #ddd;
            font-size: 14px;
        }
        button {
            background: #667eea;
            color: white;
            border: none;
            cursor: pointer;
            font-weight: bold;
        }
        button:hover {
            background: #5568d3;
        }
        .domain-list {
            list-style: none;
        }
        .domain-item {
            padding: 15px;
            background: white;
            margin-bottom: 10px;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }
        .domain-link {
            color: #667eea;
            text-decoration: none;
            font-weight: bold;
            font-size: 16px;
        }
        .domain-link:hover {
            text-decoration: underline;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }
        .stat-box {
            background: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
        }
        .stat-number {
            font-size: 32px;
            font-weight: bold;
            color: #667eea;
        }
        .stat-label {
            color: #666;
            margin-top: 5px;
        }
        .form-group {
            margin-bottom: 15px;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #333;
        }
        .form-group input {
            width: 100%;
        }
        .success {
            background: #d4edda;
            color: #155724;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            display: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 HMouth Proxy Control Panel</h1>
        <p class="subtitle">Anonymous .hmouth Domain Hosting</p>

        <div class="success" id="successMsg"></div>

        <div class="section">
            <h2>📊 Statistics</h2>
            <div class="stats" id="stats">
                <div class="stat-box">
                    <div class="stat-number" id="hostedCount">0</div>
                    <div class="stat-label">Hosted Sites</div>
                </div>
                <div class="stat-box">
                    <div class="stat-number" id="discoveredCount">0</div>
                    <div class="stat-label">Discovered Domains</div>
                </div>
                <div class="stat-box">
                    <div class="stat-number" id="peerCount">0</div>
                    <div class="stat-label">Connected Peers</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>➕ Host New Site</h2>
            
            <div style="margin-bottom: 20px;">
                <label style="margin-right: 20px;">
                    <input type="radio" name="hostType" value="static" checked onchange="toggleHostType()"> 
                    📁 Static Files
                </label>
                <label>
                    <input type="radio" name="hostType" value="backend" onchange="toggleHostType()"> 
                    🔧 Backend Application
                </label>
            </div>

            <div id="staticForm">
                <div class="form-group">
                    <label>Content Folder Path:</label>
                    <input type="text" id="contentPath" placeholder="C:\path\to\your\website">
                </div>
            </div>

            <div id="backendForm" style="display: none;">
                <div class="form-group">
                    <label>Backend URL:</label>
                    <input type="text" id="backendURL" placeholder="http://localhost:3000">
                    <small style="color: #666; display: block; margin-top: 5px;">
                        Your backend must be running on this URL
                    </small>
                </div>
            </div>

            <div class="form-group">
                <label>Custom Domain (optional):</label>
                <input type="text" id="customDomain" placeholder="mysite (will become mysite.hmouth)">
            </div>
            <button onclick="hostSite()">🚀 Host Site</button>
        </div>

        <div class="section">
            <h2>🌐 Your Hosted Domains</h2>
            <ul class="domain-list" id="hostedDomains">
                <li style="color: #666;">No sites hosted yet</li>
            </ul>
        </div>

        <div class="section">
            <h2>🔍 Discovered .hmouth Domains</h2>
            <ul class="domain-list" id="discoveredDomains">
                <li style="color: #666;">Discovering domains...</li>
            </ul>
        </div>

        <div class="section">
            <h2>⚙️ Browser Configuration</h2>
            <p><strong>Firefox:</strong></p>
            <ol>
                <li>Open Settings → Network Settings</li>
                <li>Select "Manual proxy configuration"</li>
                <li>HTTP Proxy: <code>localhost</code>, Port: <code>` + strings.TrimPrefix(hp.proxyPort, ":") + `</code></li>
                <li>Check "Also use this proxy for HTTPS"</li>
                <li>Click OK</li>
            </ol>
            <p style="margin-top: 15px;"><strong>Chrome:</strong></p>
            <ol>
                <li>Settings → System → Open proxy settings</li>
                <li>LAN Settings → Use a proxy server</li>
                <li>Address: <code>localhost</code>, Port: <code>` + strings.TrimPrefix(hp.proxyPort, ":") + `</code></li>
            </ol>
        </div>
    </div>

    <script>
        function toggleHostType() {
            const hostType = document.querySelector('input[name="hostType"]:checked').value;
            const staticForm = document.getElementById('staticForm');
            const backendForm = document.getElementById('backendForm');
            
            if (hostType === 'static') {
                staticForm.style.display = 'block';
                backendForm.style.display = 'none';
            } else {
                staticForm.style.display = 'none';
                backendForm.style.display = 'block';
            }
        }

        async function hostSite() {
            const hostType = document.querySelector('input[name="hostType"]:checked').value;
            const customDomain = document.getElementById('customDomain').value;

            let endpoint, body;

            if (hostType === 'static') {
                const contentPath = document.getElementById('contentPath').value;
                if (!contentPath) {
                    alert('Please enter a content folder path');
                    return;
                }
                endpoint = '/api/host';
                body = {contentPath, customDomain};
            } else {
                const backendURL = document.getElementById('backendURL').value;
                if (!backendURL) {
                    alert('Please enter a backend URL');
                    return;
                }
                endpoint = '/api/host-backend';
                body = {backendURL, customDomain};
            }

            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(body)
            });

            const data = await response.json();
            if (data.success) {
                const msg = document.getElementById('successMsg');
                msg.textContent = '✅ Site hosted at: ' + data.domain;
                msg.style.display = 'block';
                setTimeout(() => msg.style.display = 'none', 5000);
                loadDomains();
                loadStats();
                
                // Clear inputs
                document.getElementById('contentPath').value = '';
                document.getElementById('backendURL').value = '';
                document.getElementById('customDomain').value = '';
            } else {
                alert('Failed to host site: ' + data.error);
            }
        }

        async function loadDomains() {
            const response = await fetch('/api/domains');
            const data = await response.json();

            const hostedList = document.getElementById('hostedDomains');
            const discoveredList = document.getElementById('discoveredDomains');

            if (data.hosted && data.hosted.length > 0) {
                hostedList.innerHTML = data.hosted.map(d => 
                    '<li class="domain-item"><a href="http://' + d + '" class="domain-link">' + d + '</a></li>'
                ).join('');
            }

            if (data.discovered && data.discovered.length > 0) {
                discoveredList.innerHTML = data.discovered.map(d => 
                    '<li class="domain-item"><a href="http://' + d + '" class="domain-link">' + d + '</a></li>'
                ).join('');
            }
        }

        async function loadStats() {
            const response = await fetch('/api/stats');
            const data = await response.json();

            document.getElementById('hostedCount').textContent = data.hostedSites || 0;
            document.getElementById('discoveredCount').textContent = data.discoveredDomains || 0;
            document.getElementById('peerCount').textContent = data.peers || 0;
        }

        // Auto-refresh
        setInterval(loadDomains, 5000);
        setInterval(loadStats, 3000);
        loadDomains();
        loadStats();
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, html)
}

func (hp *HMouthProxy) handleHostSite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContentPath  string `json:"contentPath"`
		CustomDomain string `json:"customDomain"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	domain, err := hp.HostSite(req.ContentPath, req.CustomDomain)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": err == nil,
		"domain":  domain,
		"error":   func() string { if err != nil { return err.Error() }; return "" }(),
	})
}

func (hp *HMouthProxy) handleHostBackend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BackendURL   string `json:"backendURL"`
		CustomDomain string `json:"customDomain"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	domain, err := hp.HostBackend(req.BackendURL, req.CustomDomain)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": err == nil,
		"domain":  domain,
		"error":   func() string { if err != nil { return err.Error() }; return "" }(),
	})
}

func (hp *HMouthProxy) handleListDomains(w http.ResponseWriter, r *http.Request) {
	hp.mu.RLock()
	defer hp.mu.RUnlock()

	hosted := make([]string, 0, len(hp.hostedSites))
	for domain := range hp.hostedSites {
		hosted = append(hosted, domain)
	}

	discovered := make([]string, 0, len(hp.domains))
	for domain := range hp.domains {
		if _, isHosted := hp.hostedSites[domain]; !isHosted {
			discovered = append(discovered, domain)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"hosted":     hosted,
		"discovered": discovered,
	})
}

func (hp *HMouthProxy) handleStats(w http.ResponseWriter, r *http.Request) {
	hp.mu.RLock()
	hostedCount := len(hp.hostedSites)
	discoveredCount := len(hp.domains)
	hp.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"hostedSites":       hostedCount,
		"discoveredDomains": discoveredCount,
		"peers":             hp.dht.GetPeerCount(),
	})
}

func main() {
	dhtPort := flag.Int("dht", 6881, "DHT port")
	p2pPort := flag.Int("p2p", 9000, "P2P port")
	proxyPort := flag.String("proxy", ":8888", "Proxy port")
	flag.Parse()

	log.Printf("🚀 Starting HMouth Proxy...")
	log.Printf("🌐 DHT Port: %d", *dhtPort)
	log.Printf("🔌 P2P Port: %d", *p2pPort)
	log.Printf("🔗 Proxy Port: %s", *proxyPort)
	log.Printf("")

	proxy, err := NewHMouthProxy(*dhtPort, *p2pPort, *proxyPort)
	if err != nil {
		log.Fatalf("❌ Failed to start: %v", err)
	}

	log.Printf("✅ Proxy ready!")
	log.Printf("🌐 Open http://localhost%s for control panel", *proxyPort)
	log.Printf("")

	if err := proxy.StartProxy(); err != nil {
		log.Fatalf("❌ Proxy error: %v", err)
	}
}
