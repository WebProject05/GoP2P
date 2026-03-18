# P2P File Sharing and Communication Tool

A decentralized peer-to-peer (P2P) system built in Go for secure communication and direct file transfer without relying on a central server.

---

## Overview

This project enables users to:

- Discover peers across a network
- Establish secure, encrypted connections
- Communicate via real-time group and private chat
- Transfer files directly between devices

The system is designed around the principle of **data sovereignty**, ensuring that user data is never processed or stored by a central authority.

---

## Core Philosophy

Unlike traditional applications where communication is routed through centralized servers, this system uses a **signaling mechanism only for peer discovery**.

Once peers are introduced, all communication happens through **direct, encrypted connections** between devices.

---

## How It Works

The system is composed of three main layers:

---

### 1. Discovery and Signaling

Devices on the internet are often behind NAT (Network Address Translation), making direct discovery difficult.

A signaling server (Matchmaker) is used to facilitate initial peer discovery.

**Process:**

- **Registration**  
  The client registers its public IP address and port with the signaling server.

- **Peer Exchange**  
  The server returns a list of active peers in the same room.

- **Hole Punching**  
  Clients attempt to establish direct connections using NAT traversal techniques.

---

### 2. Cryptographic Handshake

All communication is secured using strong cryptographic protocols.

**Steps:**

- **Key Exchange (ECDH)**  
  Peers exchange public keys using Elliptic-Curve Diffie-Hellman.

- **Shared Secret Generation**  
  Each peer derives a shared secret using their private key and the peer’s public key.

- **Encryption (AES-256-GCM)**  
  All data is encrypted using AES-GCM with the shared secret.  
  The secret is never transmitted over the network.

---

### 3. Mesh Networking

In a common room, the system forms a decentralized mesh network.

- Each peer maintains direct connections with all other peers
- No central host or coordinator
- Messages are sent individually to each peer

For example:
- In a room of 5 users, each node maintains 4 connections

---

## Features

- Decentralized peer-to-peer communication
- Encrypted group chat (LAN and internet)
- Direct 1-on-1 encrypted messaging
- Secure file transfer with progress tracking
- No central data storage
- Ephemeral messaging (no chat persistence)

---

## Commands

### Join Room
```bash
p2p room <username> <signaler_ip>
