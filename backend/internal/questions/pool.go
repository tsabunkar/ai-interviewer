package questions

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/ai-interviewer/backend/internal/models"
)

// questionPool is a curated set of system design interview questions used as
// a deterministic fallback when Bedrock question generation fails.
var questionPool = []models.Question{
	{
		QuestionID: "sd-twitter",
		Title:      "Design Twitter",
		Description: `## Design Twitter (X)

Design a social media platform similar to Twitter that supports hundreds of millions of users.

### Functional Requirements
- Users can post tweets (up to 280 characters) with optional media attachments
- Users can follow/unfollow other users
- Users have a home timeline showing tweets from people they follow, in reverse chronological order
- Users can like, retweet, and reply to tweets
- Support @mentions and hashtags with search capability

### Non-Functional Requirements
- The system should be highly available (99.99% uptime)
- Timeline generation should have low latency (< 200ms for p99)
- The system should handle 500M+ daily active users
- Eventual consistency is acceptable for timeline delivery

### Constraints
- Peak tweet rate: ~10,000 tweets/second
- Average user follows 200 people
- Celebrity users may have 50M+ followers (fan-out problem)
- Read-heavy system: read-to-write ratio of 1000:1

### What We Expect
- A clear data model for users, tweets, and relationships
- A feed generation strategy (fan-out on write vs fan-out on read or hybrid)
- Media storage and CDN strategy
- Caching layer design
- Search indexing approach`,
		Difficulty: "Hard",
		Categories: []string{"Database", "Cache", "Scalability", "API Design", "Availability"},
		Hints: []string{
			"Consider a hybrid fan-out approach: fan-out on write for normal users, fan-out on read for celebrity users",
			"Think about using a combination of Redis for timelines and a NoSQL store for tweet storage",
			"Consider how you would handle the thundering herd problem when a celebrity tweets",
		},
	},
	{
		QuestionID: "sd-url-shortener",
		Title:      "Design URL Shortener",
		Description: `## Design a URL Shortener (like bit.ly)

Design a URL shortening service that converts long URLs into short, unique aliases.

### Functional Requirements
- Given a long URL, generate a short unique URL
- When users access a short URL, redirect them to the original long URL
- Users can optionally set custom short URLs
- URLs should expire after a configurable time period
- Analytics: track click count, geographic distribution, referrer info

### Non-Functional Requirements
- The system should be highly available; URL redirection must work 24/7
- URL redirection should be real-time with minimal latency (< 10ms)
- Short URLs should not be predictable (no sequential IDs)

### Constraints
- 100M new URLs shortened per month
- 10:1 read-to-write ratio (1B redirections per month)
- Short URL length: 7 characters using [a-zA-Z0-9]
- Data retention: 5 years by default
- Average long URL size: 500 bytes

### What We Expect
- URL encoding/hashing strategy and collision handling
- Database schema and choice of database
- Caching strategy for hot URLs
- Rate limiting to prevent abuse
- 301 vs 302 redirect trade-offs`,
		Difficulty: "Easy",
		Categories: []string{"Database", "Cache", "API Design", "Scalability"},
		Hints: []string{
			"Base62 encoding of an auto-incrementing ID or MD5/SHA256 hash can be used for URL generation",
			"Consider using a bloom filter for quick collision detection",
			"Think about cache eviction policies — LRU works well since recently created URLs are accessed most",
		},
	},
	{
		QuestionID: "sd-netflix",
		Title:      "Design Netflix",
		Description: `## Design Netflix

Design a video streaming platform that serves millions of concurrent viewers worldwide.

### Functional Requirements
- Users can browse, search, and stream video content
- Support multiple video quality levels with adaptive bitrate streaming
- Personalized recommendations based on viewing history
- User profiles, watch lists, and viewing history
- Content upload and transcoding pipeline for content providers

### Non-Functional Requirements
- High availability: 99.99% uptime for streaming
- Low startup latency: video should begin playing within 2 seconds
- Support 200M+ subscribers globally
- Handle peak concurrent streams of 10M+

### Constraints
- Video catalog: 15,000+ titles
- Each title stored in multiple resolutions (360p to 4K) and codecs
- Average video size: 3GB (compressed, single resolution)
- Global content delivery required across 190+ countries
- DRM (Digital Rights Management) is required

### What We Expect
- Video storage and CDN architecture
- Adaptive bitrate streaming protocol choice (HLS/DASH)
- Content transcoding pipeline design
- Recommendation engine high-level architecture
- User session management and concurrent stream limiting`,
		Difficulty: "Hard",
		Categories: []string{"Scalability", "Availability", "Database", "Cache", "Tradeoffs"},
		Hints: []string{
			"Consider using a microservices architecture with separate services for user management, content catalog, streaming, and recommendations",
			"Think about how Netflix uses Open Connect — their custom CDN with ISP-embedded appliances",
			"Adaptive bitrate streaming (ABR) is key — the client dynamically adjusts quality based on bandwidth",
		},
	},
	{
		QuestionID: "sd-uber",
		Title:      "Design Uber",
		Description: `## Design Uber

Design a ride-sharing platform that matches riders with nearby drivers in real-time.

### Functional Requirements
- Riders can request rides by specifying pickup and drop-off locations
- System matches the nearest available driver to a rider
- Real-time tracking of driver location during the trip
- Fare estimation before ride and final fare calculation after ride
- Rating system for both drivers and riders
- Support for different ride types (UberX, UberXL, UberBlack)

### Non-Functional Requirements
- Low latency for driver matching (< 5 seconds)
- Real-time location updates (every 3-5 seconds)
- High availability in all service cities
- Handle 15M+ daily trips globally

### Constraints
- 5M+ active drivers
- 100M+ active riders
- Location update frequency: every 4 seconds per active driver
- Peak requests: 50,000 ride requests per minute
- Geographic coverage: 10,000+ cities

### What We Expect
- Location-based matching algorithm (geospatial indexing)
- Real-time location tracking architecture
- Supply-demand based pricing (surge pricing)
- Trip lifecycle management (request → match → pickup → ride → dropoff)
- ETA calculation approach`,
		Difficulty: "Hard",
		Categories: []string{"Scalability", "Database", "API Design", "Availability", "Tradeoffs"},
		Hints: []string{
			"Consider using geospatial indexes like Geohash or Quadtrees for efficient proximity searches",
			"Think about using WebSockets or Server-Sent Events for real-time location updates",
			"A dispatch service with a supply positioning algorithm can optimize driver-rider matching",
		},
	},
	{
		QuestionID: "sd-whatsapp",
		Title:      "Design WhatsApp",
		Description: `## Design WhatsApp

Design a real-time messaging application that supports billions of users.

### Functional Requirements
- One-on-one messaging with text, images, video, and documents
- Group messaging (up to 1024 members)
- Online/offline status and last seen timestamps
- Message delivery receipts (sent, delivered, read)
- End-to-end encryption for all messages
- Message history sync across devices

### Non-Functional Requirements
- Real-time message delivery (< 100ms for online users)
- Messages must not be lost — at-least-once delivery guarantee
- Support 2B+ registered users, 100M+ daily active users
- High availability with minimal downtime

### Constraints
- 100B+ messages sent per day
- Average message size: 100 bytes (text), up to 16MB (media)
- Users are on multiple devices simultaneously
- Messages stored for 30 days on server (for offline delivery)
- Group size limit: 1024 members

### What We Expect
- Connection management (WebSocket vs long polling)
- Message delivery and acknowledgment protocol
- Offline message queuing and synchronization
- Group messaging fan-out strategy
- End-to-end encryption key management overview
- Media storage and delivery architecture`,
		Difficulty: "Hard",
		Categories: []string{"Scalability", "Availability", "Database", "API Design", "Tradeoffs"},
		Hints: []string{
			"Consider using a message queue per user for reliable delivery of offline messages",
			"Think about the Signal Protocol for end-to-end encryption with perfect forward secrecy",
			"For group messages, consider writing once and having each recipient's connection server fetch on demand",
		},
	},
	{
		QuestionID: "sd-instagram",
		Title:      "Design Instagram",
		Description: `## Design Instagram

Design a photo and video sharing social network.

### Functional Requirements
- Users can upload photos and short videos with captions
- Users can follow other users
- News feed showing posts from followed users in ranked order
- Like and comment on posts
- Explore/discover page with trending content
- Stories feature (ephemeral content that disappears after 24 hours)

### Non-Functional Requirements
- High availability: the service should always be accessible
- News feed generation latency < 500ms
- Photo upload should complete within 5 seconds
- Support 1B+ monthly active users

### Constraints
- 100M+ photos uploaded daily
- Average photo size: 2MB, average video: 50MB
- Storage grows by ~200TB per day
- Read-heavy: 100:1 read-to-write ratio
- Global user base requiring low-latency access

### What We Expect
- Photo/video storage and CDN strategy
- News feed generation algorithm (ranked vs chronological)
- Database schema design for users, posts, follows, likes
- Sharding strategy for the social graph
- Content moderation pipeline overview`,
		Difficulty: "Medium",
		Categories: []string{"Database", "Cache", "Scalability", "API Design"},
		Hints: []string{
			"Consider separating photo metadata from the actual photo storage (object store + metadata DB)",
			"Pre-compute news feeds for active users using a fan-out-on-write approach",
			"Think about using a graph database or adjacency list for the follower relationship",
		},
	},
	{
		QuestionID: "sd-youtube",
		Title:      "Design YouTube",
		Description: `## Design YouTube

Design a video sharing and streaming platform at global scale.

### Functional Requirements
- Users can upload videos of various formats and lengths
- Users can stream/watch videos with adaptive quality
- Search videos by title, description, tags
- Like, dislike, comment on videos
- Subscribe to channels and receive notifications
- Video recommendations based on watch history

### Non-Functional Requirements
- Videos should start playing within 2 seconds
- Support 2B+ logged-in users per month
- 99.9% availability for video playback
- Smooth streaming with minimal buffering

### Constraints
- 500+ hours of video uploaded per minute
- 1B+ hours of video watched daily
- Videos stored in multiple resolutions (144p to 8K)
- Global audience requiring edge caching
- Average video size varies from 10MB to 50GB

### What We Expect
- Video upload and transcoding pipeline
- Video storage strategy (chunking, multiple resolutions)
- Content delivery network design
- Search and recommendation architecture
- View count and analytics at scale (near real-time)`,
		Difficulty: "Hard",
		Categories: []string{"Scalability", "Availability", "Database", "Cache", "Tradeoffs"},
		Hints: []string{
			"Consider a DAG-based transcoding pipeline where each resolution is processed in parallel",
			"Think about chunked uploads with resumability for large video files",
			"Use approximate counting (like HyperLogLog) for view counts to avoid write hotspots",
		},
	},
	{
		QuestionID: "sd-dropbox",
		Title:      "Design Dropbox",
		Description: `## Design Dropbox

Design a cloud file storage and synchronization service.

### Functional Requirements
- Users can upload, download, and delete files from any device
- Automatic file synchronization across all user devices
- File versioning — ability to view and restore previous versions
- File and folder sharing with other users (view/edit permissions)
- Offline access with sync when connectivity is restored

### Non-Functional Requirements
- High reliability: files must never be lost (99.999999999% durability)
- Real-time sync: changes reflected across devices within seconds
- Low bandwidth usage through delta sync
- Support 500M+ registered users

### Constraints
- Average user has 2GB of files; power users up to 2TB
- Total storage: hundreds of petabytes
- Support files up to 50GB
- Concurrent editing by multiple users on shared files
- Users have 3+ connected devices on average

### What We Expect
- File chunking and deduplication strategy
- Delta sync algorithm for bandwidth efficiency
- Conflict resolution for concurrent edits
- Metadata database design vs block storage architecture
- Notification system for real-time sync across devices`,
		Difficulty: "Medium",
		Categories: []string{"Database", "Scalability", "Availability", "API Design", "Tradeoffs"},
		Hints: []string{
			"Consider chunking files into 4MB blocks and using content-based hashing for deduplication",
			"Think about using a sync protocol based on Merkle trees for efficient diff detection",
			"Separate the metadata service (SQL) from the block storage service (object store like S3)",
		},
	},
	{
		QuestionID: "sd-rate-limiter",
		Title:      "Design a Rate Limiter",
		Description: `## Design a Rate Limiter

Design a distributed rate limiting system that can be used to protect APIs and services from abuse.

### Functional Requirements
- Limit the number of requests a client can make within a time window
- Support multiple rate limiting strategies: fixed window, sliding window, token bucket
- Allow different rate limits per API endpoint, user tier, or client
- Return appropriate HTTP 429 responses with retry-after headers
- Provide rate limit headers in every response (X-RateLimit-Limit, X-RateLimit-Remaining)

### Non-Functional Requirements
- Ultra-low latency: rate limit check must complete in < 1ms
- High availability: if the rate limiter is down, default to allowing traffic (fail-open)
- Distributed: work consistently across multiple server instances
- Accurate: minimal over-counting or under-counting

### Constraints
- Handle 10M+ requests per second across the platform
- Support 1M+ distinct clients
- Rate limit rules change infrequently (can be cached)
- Must work in a multi-region deployment
- Memory-efficient: can't store per-request data

### What We Expect
- Comparison of rate limiting algorithms with trade-offs
- Distributed counter management (Redis-based approach)
- Race condition handling in distributed environments
- Configuration management for rate limit rules
- Client identification strategy (API key, IP, user ID)`,
		Difficulty: "Medium",
		Categories: []string{"Cache", "Scalability", "API Design", "Availability", "Tradeoffs"},
		Hints: []string{
			"The sliding window log algorithm is most accurate but memory-intensive; sliding window counter is a good compromise",
			"Consider using Redis with Lua scripts for atomic increment-and-check operations",
			"Think about how to handle rate limiting in a multi-datacenter setup — local vs global counters",
		},
	},
	{
		QuestionID: "sd-web-crawler",
		Title:      "Design a Web Crawler",
		Description: `## Design a Web Crawler

Design a scalable web crawler that can crawl billions of web pages efficiently.

### Functional Requirements
- Crawl web pages starting from a set of seed URLs
- Extract and follow links to discover new pages
- Store crawled content for indexing and analysis
- Respect robots.txt rules and crawl politeness policies
- Detect and handle duplicate content
- Support recrawling to keep content fresh

### Non-Functional Requirements
- Scalability: crawl 1B+ pages per week
- Politeness: do not overwhelm any single web server
- Robustness: handle malformed HTML, infinite loops, spider traps
- Extensibility: easy to add new content processors

### Constraints
- Average page size: 500KB (including resources)
- Billions of URLs to manage in the frontier
- Pages need recrawling based on change frequency
- Must handle various content types (HTML, PDF, images)
- DNS resolution bottleneck at scale

### What We Expect
- URL frontier design with priority and politeness queues
- Distributed crawling architecture
- Content deduplication strategy (URL-level and content-level)
- Storage design for raw content and parsed data
- Crawl scheduling and freshness management`,
		Difficulty: "Medium",
		Categories: []string{"Scalability", "Database", "Tradeoffs"},
		Hints: []string{
			"Use a URL frontier with separate queues per domain to enforce politeness (one request per domain at a time)",
			"Consider using consistent hashing to assign URL domains to specific crawler instances",
			"SimHash or MinHash can efficiently detect near-duplicate content without storing full page content",
		},
	},
	{
		QuestionID: "sd-notification",
		Title:      "Design a Notification System",
		Description: `## Design a Notification System

Design a scalable notification system that supports multiple channels and millions of users.

### Functional Requirements
- Support multiple notification channels: push (iOS/Android), SMS, email, in-app
- User notification preferences (opt-in/opt-out per channel and category)
- Template-based notifications with variable substitution
- Notification scheduling (send immediately or at a specified time)
- Notification history and read/unread status tracking
- Rate limiting to prevent notification fatigue

### Non-Functional Requirements
- High throughput: send 10M+ notifications per day
- Low latency for real-time notifications (< 5 seconds)
- At-least-once delivery guarantee
- Graceful degradation when downstream providers are unavailable

### Constraints
- Multiple third-party providers per channel (fallback support)
- Template library with 1000+ templates
- User preferences must be checked before every send
- Compliance with regulations (CAN-SPAM, GDPR)
- Support for localization (50+ languages)

### What We Expect
- Message queue architecture for reliable delivery
- Channel-specific delivery pipelines
- User preference service design
- Template rendering engine
- Retry and dead-letter queue strategy
- Analytics and delivery tracking`,
		Difficulty: "Medium",
		Categories: []string{"Scalability", "Availability", "API Design", "Database", "Tradeoffs"},
		Hints: []string{
			"Consider using separate message queues per notification channel for independent scaling",
			"Think about implementing a priority queue so critical notifications (OTP, security alerts) are sent first",
			"Use a circuit breaker pattern for third-party provider integrations to handle failures gracefully",
		},
	},
	{
		QuestionID: "sd-payment",
		Title:      "Design a Payment System",
		Description: `## Design a Payment System

Design a payment processing system like Stripe or PayPal that handles financial transactions securely and reliably.

### Functional Requirements
- Process payments: authorize, capture, refund, void
- Support multiple payment methods: credit card, debit card, bank transfer, digital wallets
- Transaction history and receipt generation
- Webhook notifications for payment status updates
- Multi-currency support with exchange rate management
- Recurring payments and subscription billing

### Non-Functional Requirements
- Exactly-once payment processing (no double charges)
- High availability: 99.999% uptime for payment processing
- PCI DSS compliance for handling card data
- Low latency: payment authorization within 2 seconds
- Strong consistency for financial data

### Constraints
- Process 1M+ transactions per day
- Average transaction size: $50
- Must handle peak loads during events (Black Friday: 10x normal)
- Multi-region deployment for regulatory compliance
- Audit trail required for every transaction

### What We Expect
- Idempotency mechanism to prevent double payments
- Payment state machine design
- Ledger/double-entry bookkeeping system
- Integration with payment service providers (PSPs)
- Reconciliation process design
- Security and fraud detection overview`,
		Difficulty: "Hard",
		Categories: []string{"Database", "Availability", "Scalability", "API Design", "Tradeoffs"},
		Hints: []string{
			"Use idempotency keys for all payment operations to guarantee exactly-once processing",
			"Consider a double-entry ledger where every transaction creates both a debit and credit entry",
			"Think about using the Saga pattern for distributed transactions across multiple services",
		},
	},
	{
		QuestionID: "sd-autocomplete",
		Title:      "Design Search Autocomplete",
		Description: `## Design Search Autocomplete

Design a real-time search autocomplete system (typeahead) that provides suggestions as users type.

### Functional Requirements
- Return top-k search suggestions as the user types each character
- Suggestions ranked by popularity (search frequency)
- Support for personalized suggestions based on user history
- Filter out inappropriate or banned queries
- Multi-language support

### Non-Functional Requirements
- Ultra-low latency: suggestions must appear within 100ms
- High availability: degraded experience without autocomplete is acceptable but not preferred
- Suggestions should be updated with trending queries within minutes
- Handle 100K+ queries per second

### Constraints
- Vocabulary: 5B+ unique search queries
- Top 20% of queries account for 80% of search volume
- Average query length: 4-5 words
- Prefix matching only (not substring matching)
- Suggestion list size: 5-10 results

### What We Expect
- Trie-based data structure design with frequency tracking
- Data collection and aggregation pipeline
- Caching strategy for popular prefixes
- Approach for real-time trending query incorporation
- Serving layer architecture for low-latency lookups`,
		Difficulty: "Medium",
		Categories: []string{"Cache", "Database", "Scalability", "API Design"},
		Hints: []string{
			"Consider using a Trie where each node stores the top-k suggestions to avoid traversal at query time",
			"Think about using a MapReduce pipeline to periodically rebuild the Trie from search logs",
			"A two-level cache (browser/client cache + server-side cache) can dramatically reduce backend load",
		},
	},
	{
		QuestionID: "sd-distributed-cache",
		Title:      "Design a Distributed Cache",
		Description: `## Design a Distributed Cache

Design a distributed caching system like Memcached or Redis that operates across multiple nodes.

### Functional Requirements
- Key-value store with GET, PUT, DELETE operations
- Support TTL (time-to-live) for automatic key expiration
- Support multiple eviction policies (LRU, LFU, FIFO)
- Data replication for fault tolerance
- Support for data types beyond simple strings (lists, sets, sorted sets)

### Non-Functional Requirements
- Sub-millisecond latency for read and write operations
- Linear horizontal scalability (add nodes to increase capacity)
- High availability: survive node failures without data loss
- Consistent hashing for minimal data movement during scaling

### Constraints
- Billions of key-value pairs in total
- Individual value size: up to 1MB
- Cluster size: 100-1000 nodes
- Network partitions must be handled gracefully
- Memory is the primary storage medium

### What We Expect
- Data partitioning strategy (consistent hashing with virtual nodes)
- Replication and consistency model (eventual vs strong)
- Cache eviction algorithm implementation
- Node discovery and failure detection (gossip protocol)
- Client-side vs server-side request routing
- Hot key handling strategy`,
		Difficulty: "Medium",
		Categories: []string{"Cache", "Scalability", "Availability", "Tradeoffs"},
		Hints: []string{
			"Consistent hashing with virtual nodes ensures uniform data distribution and minimal reshuffling",
			"Consider the CAP theorem trade-offs: Redis Cluster chooses AP, while some systems choose CP",
			"For hot keys, consider local caching on the client side or key replication across multiple nodes",
		},
	},
	{
		QuestionID: "sd-code-editor",
		Title:      "Design an Online Code Editor",
		Description: `## Design an Online Code Editor

Design a collaborative online code editor (like Replit, CodeSandbox, or Google Colab) that supports real-time collaboration and code execution.

### Functional Requirements
- Browser-based code editor with syntax highlighting and IntelliSense
- Real-time collaborative editing (multiple users editing simultaneously)
- Code execution in multiple languages (Python, JavaScript, Go, Java, C++)
- Project file management (create, rename, delete files and folders)
- Terminal/console output display
- Version history and ability to fork projects

### Non-Functional Requirements
- Real-time sync latency < 100ms for collaborative editing
- Code execution should start within 2 seconds
- Support 1M+ concurrent active sessions
- Isolation between code execution environments (security)

### Constraints
- Code execution timeout: 30 seconds per run
- Memory limit per execution: 256MB
- Maximum project size: 100MB
- Support 20+ programming languages
- Concurrent collaborators per project: up to 10

### What We Expect
- Real-time collaboration protocol (OT vs CRDT)
- Sandboxed code execution architecture (containers, VMs, or WebAssembly)
- File system abstraction and storage
- WebSocket architecture for real-time features
- Session management and conflict resolution`,
		Difficulty: "Hard",
		Categories: []string{"Scalability", "Availability", "API Design", "Database", "Tradeoffs"},
		Hints: []string{
			"CRDTs (Conflict-free Replicated Data Types) are increasingly preferred over OT for collaborative editing due to their simpler consistency model",
			"Consider using containerized sandboxes (like gVisor or Firecracker) for secure code execution",
			"Think about using a virtual file system that syncs with object storage for project persistence",
		},
	},
}

// GetRandomQuestion returns a question from the curated pool, deterministically
// selected based on the provided date string. This ensures the same date always
// yields the same question, providing consistency across retries.
func GetRandomQuestion(date string) models.Question {
	h := sha256.Sum256([]byte(date))
	idx := int(binary.BigEndian.Uint64(h[:8])) % len(questionPool)
	if idx < 0 {
		idx = -idx
	}

	q := questionPool[idx]
	q.Date = date
	return q
}
