# 🔒 HashMouth - Anonymous P2P Communication

HashMouth is a decentralized, anonymous communication platform with DHT peer discovery, .hmouth domains (like Tor's .onion), and multiple anonymity layers.

### Host on .hmouth Domain
```bash
# Start proxy
go run cmd/hmouth_proxy.go

# Open control panel: http://localhost:8888
# Host your backend as chat.hmouth
# Configure browser proxy: localhost:8888
# Visit: http://chat.hmouth
```

## 🌟 Features

### 1. **DHT Peer Discovery**
- Automatic peer discovery using BitTorrent DHT
- No central server needed
- Connects to public DHT bootstrap nodes
- Works like torrent peer discovery

## 📋 What You Can Do

### 1. Simple Chat
```bash
cd examples/simple_backend
npm install
npm start
# Open: http://localhost:3000
```
- Enter your name
- Chat with everyone or specific users
- Simple and fast

### 2. Host .hmouth Websites
```bash
go run cmd/hmouth_proxy.go
```
- Host static sites or backends
- Access via .hmouth domains
- Anonymous hosting
- Like Tor hidden services

### 4. DHT Chat
```bash
go run cmd/dht_chat.go
```
- Automatic peer discovery
- No manual setup
- DHT-based

## 🏗️ Architecture

### Core Components

**DHT Network** (`network/dht.go`)
- BitTorrent DHT implementation
- Connects to public bootstrap nodes
- Automatic peer discovery
- Announces presence

**P2P Network** (`network/node.go`)
- TCP-based P2P connections
- Encrypted communication
- Peer management

**Relay Network** (`network/relay.go`)
- Multi-hop message routing
- Anonymous communication
- Path building

**Crypto** (`crypto/`)
- ChaCha20-Poly1305 encryption
- Onion packet creation
- Key managemen

## 🔧 Requirements

**Go Applications:**
- Go 1.23 or higher
- golang.org/x/crypto

**Node.js Backend:**
- Node.js 14+
- npm packages: express, cors

## 🆚 Comparison

| Feature | Simple Chat | .hmouth | Ultra Anonymous | DHT Chat |
|---------|-------------|---------|-----------------|----------|
| Setup | Easy | Medium | Medium | Easy |
| Anonymity | Low | Medium | Maximum | Medium |
| Speed | Fast | Medium | Slow | Fast |
| Discovery | Manual | DHT | Manual | DHT |
| Best For | Friends | Hosting | Privacy | Public |

## 🤝 Contributing

Contributions welcome! This is an experimental project.

## ⚠️ Disclaimer

This is experimental software. Not audited for production use. Use at your own risk.

## 📄 License

See LICENSE file for details.

---

**Made with ❤️ for anonymous communication**
