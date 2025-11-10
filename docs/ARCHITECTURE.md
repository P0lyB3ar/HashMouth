# HashMouth Architecture

## Overview

HashMouth is a secure, anonymous messaging system that combines multiple privacy-preserving techniques:

- **Onion Routing**: Multi-layer encryption for anonymous message routing
- **Mix Networks**: Traffic analysis resistance through batching and delays
- **End-to-End Encryption**: ChaCha20-Poly1305 authenticated encryption
- **Message Chunking**: Support for large messages
- **Double Ratchet**: Forward secrecy for session keys

## System Components

### 1. Crypto Layer (`crypto/`)

The cryptographic foundation of HashMouth.

#### crypto.go
- **OnionPacket**: Encrypted data container
- **CreateOnionPacket()**: Encrypts data with ChaCha20-Poly1305
- **PeelOnion()**: Decrypts one layer of encryption
- **Serialize/Deserialize**: Packet serialization

#### keys.go
- **GenerateIdentityKeyPair()**: Creates Ed25519 keypairs for node identity
- **GenerateSymmetricKey()**: Creates 32-byte keys for ChaCha20-Poly1305

#### ratchet.go
- **RatchetSession**: Manages session state with a peer
- **NewRatchetSession()**: Initializes session with X25519 key exchange
- **RatchetStep()**: Evolves chain key for forward secrecy
- **GetNextKey()**: Returns key for next message

### 2. Message Layer (`message/`)

Handles message formatting and fragmentation.

#### chunk.go
- **Chunk**: Represents a message fragment
- **SplitMessage()**: Splits large messages into chunks
- **ChunkAssembler**: Reassembles chunks into complete messages
- **Validate()**: Ensures chunk integrity

#### packet.go
- **Packet**: Network packet with metadata
- **PacketType**: Defines packet types (Data, Ack, Handshake, KeyExchange)
- **Sign()**: Signs packet with Ed25519
- **Verify()**: Verifies packet signature
- **IsExpired()**: Checks for replay attacks

### 3. Routing Layer (`routing/`)

Manages path selection and mix network operations.

#### path.go
- **Path**: Ordered list of nodes for routing
- **PathBuilder**: Constructs random paths through the network
- **BuildRandomPath()**: Creates random path with specified length
- **BuildPathExcluding()**: Creates path avoiding certain nodes
- **BuildMultiplePaths()**: Creates multiple diverse paths

#### mixnode.go
- **MixNode**: Implements mix network node
- **AddPacket()**: Queues packet for processing
- **processBatch()**: Batches and shuffles packets
- **randomDelay()**: Adds timing obfuscation
- **MixNetwork**: Manages multiple mix nodes

### 4. Network Layer (`network/`)

Provides P2P networking primitives.

#### node.go
- **P2PNode**: Peer-to-peer network node
- **Listen()**: Starts TCP listener
- **ConnectPeer()**: Establishes connection to peer
- **SendMessage()**: Sends data to peer
- **handleConn()**: Handles incoming connections

## Message Flow

### Sending a Message

1. **Message Creation**
   - User creates plaintext message
   - Message is split into chunks if large

2. **Encryption**
   - Each chunk is encrypted with recipient's key
   - Session key is derived using ratchet

3. **Onion Wrapping**
   - Select path through network
   - Encrypt message for each hop (innermost to outermost)
   - Each layer contains next hop information

4. **Mix Network**
   - Packet enters mix node
   - Batched with other packets
   - Shuffled and delayed
   - Forwarded to next hop

5. **Transmission**
   - Packet sent over TCP
   - Each node peels one layer
   - Forwards to next hop

### Receiving a Message

1. **Reception**
   - Node receives encrypted packet
   - Verifies packet signature

2. **Decryption**
   - Peels onion layer with node's key
   - Determines if final recipient or relay

3. **Relay or Deliver**
   - If relay: forward to next hop
   - If final: decrypt with session key

4. **Reassembly**
   - Collect all chunks
   - Verify completeness
   - Reassemble original message

## Security Properties

### Anonymity
- **Sender Anonymity**: Mix network hides sender
- **Recipient Anonymity**: Multiple paths obscure destination
- **Relationship Anonymity**: Traffic analysis resistance

### Confidentiality
- **End-to-End Encryption**: Only sender and recipient can read
- **Forward Secrecy**: Compromised keys don't reveal past messages
- **Layer Encryption**: Each hop only sees next destination

### Integrity
- **Message Authentication**: Ed25519 signatures
- **Packet Integrity**: ChaCha20-Poly1305 AEAD
- **Replay Protection**: Timestamps and nonces

## Performance Considerations

### Latency
- Mix node delays: 50-200ms per hop
- Path length: 2-4 hops typical
- Total latency: 100-800ms

### Throughput
- Chunk size: Configurable (default ~1KB)
- Batch size: 5-10 packets per batch
- Network overhead: ~20% for encryption/signatures

### Scalability
- Nodes: Designed for 100s of nodes
- Concurrent connections: Limited by OS
- Message size: Unlimited (via chunking)

## Future Enhancements

1. **Directory Service**: Automated peer discovery
2. **Cover Traffic**: Dummy packets for traffic analysis resistance
3. **Group Messaging**: Efficient multi-recipient encryption
4. **Mobile Support**: Optimizations for mobile devices
5. **Persistence**: Message storage and offline delivery
6. **DHT Integration**: Decentralized routing tables
7. **Post-Quantum Crypto**: Quantum-resistant algorithms

## References

- [Tor: The Second-Generation Onion Router](https://svn.torproject.org/svn/projects/design-paper/tor-design.pdf)
- [The Signal Protocol](https://signal.org/docs/)
- [Mixminion: Design of a Type III Anonymous Remailer Protocol](https://www.mixminion.net/)
- [ChaCha20-Poly1305 AEAD](https://tools.ietf.org/html/rfc8439)
