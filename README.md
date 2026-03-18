# Resrve

**A high-integrity reservation and coordination service with expiring ownership and conflict guarantees.**

---

## 🚀 Overview

Resrve is a backend service that enables **safe, time-bound ownership of resources** in concurrent and distributed environments.

It ensures that:

* Only one client can hold a resource at a time
* Ownership automatically expires if not renewed
* Stale clients can never act on expired ownership
* All operations are safe under retries, delays, and failures

At its core, Resrve implements a **lease-based coordination model** with **versioned ownership (fencing tokens)**.

---

## 💡 Why This Exists

Many real-world systems require temporary ownership of shared resources:

* Reserving a seat before purchase
* Assigning jobs to workers
* Electing a leader in a cluster
* Preventing duplicate actions (e.g. payments)

Naive implementations often lead to:

* race conditions
* double allocation
* stale clients performing invalid operations

Resrve solves these problems by treating **time and ownership as first-class concepts**.

---

## 🧠 Core Concepts

### Lease (Reservation)

A lease grants exclusive ownership of a resource for a limited time (TTL).

### Expiration

If the lease is not renewed before its TTL expires, it is automatically released.

### Fencing Token (Version)

Every lease operation produces a **monotonically increasing token**.

This guarantees:

> Old clients can never overwrite or interfere with newer ownership.

### Idempotency

Operations are designed to be safely retried without causing inconsistencies.

---

## 🌐 API

### Acquire a Lease

```http
POST /leases/{resource}
```

Response:

```json
{
  "lease_id": "abc123",
  "holder": "client1",
  "token": 7,
  "expires_at": "2026-03-01T12:34:56Z"
}
```

---

### Renew a Lease

```http
PUT /leases/{lease_id}/renew
```

* Extends the lease duration
* Generates a new token

---

### Release a Lease

```http
DELETE /leases/{lease_id}
```

* Frees the resource immediately

---

### Get Lease State

```http
GET /leases/{resource}
```

---

## ⚙️ Guarantees

Resrve provides the following guarantees:

* **Mutual Exclusion**
  Only one active lease per resource

* **Time-Bound Ownership**
  Leases expire automatically

* **No Stale Writes**
  Fencing tokens prevent outdated clients from acting

* **Retry Safety**
  Requests can be repeated without corrupting state

---

## 🔥 Example Scenario

Two clients attempt to reserve the same resource:

1. Client A acquires lease → token = 1
2. Client A disconnects (does not release)
3. Lease expires
4. Client B acquires lease → token = 2

If Client A later tries to act using token = 1:

❌ Operation is rejected (stale token)

---

## 🧪 Use Cases

* Distributed locks
* Job processing systems
* Leader election
* Reservation systems (tickets, inventory, etc.)
* Rate limiting coordination
* Preventing duplicate actions

---

## 🛠️ Implementation Highlights

* In-memory lease store with concurrency control
* TTL-based expiration with background cleanup
* Monotonic token generation per resource
* Strict validation of ownership and expiration
* HTTP API with clear semantics and error handling

---

## 📚 What This Project Demonstrates

* Distributed systems fundamentals
* Correctness under concurrency and failure
* Time-based coordination (TTL, expiration)
* API design for safety and idempotency
* Real-world coordination primitives used in production systems

---

## 🚧 Future Work

* Persistent storage (Redis / database)
* Multi-node coordination
* Watch/subscribe API for lease changes
* Fair queuing and priority handling

---

## 📌 Summary

Resrve is not just a reservation system.

It is a **general-purpose coordination primitive** designed to handle **ownership, time, and correctness** in distributed environments.

---

## 📖 License

MIT
