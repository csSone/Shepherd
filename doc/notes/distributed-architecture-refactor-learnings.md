# Distributed Architecture Refactor - Main.go Refactoring

## Date: 2026-02-20

## Summary
Successfully refactored cmd/shepherd/main.go to support the new distributed architecture with role-based initialization.

## Changes Made

### 1. New App Struct
Created an App struct to encapsulate all application components:
- Base components: config, process manager, model manager, shutdown manager, server
- Distributed components: node, nodeManager, masterHandler, masterConnector
- Role tracking: role string (standalone/master/client/hybrid)

### 2. Role-Based Initialization
Implemented determineRole() method that:
- Prioritizes cfg.Node.Role (new architecture)
- Falls back to cfg.Mode (backward compatibility)
- Supports: standalone, master, client, hybrid roles

### 3. Four Operating Modes

#### Standalone Mode
- Creates local Node (optional)
- Maintains full backward compatibility
- No distributed components

#### Master Mode
- Creates Node with Master role
- Initializes NodeManager for client node management
- Creates MasterHandler for HTTP API routes
- No MasterConnector (not connecting to upstream)

#### Client Mode
- Creates Node with Client role
- Initializes MasterConnector for upstream connection
- Includes heartbeat management and command execution
- Fails if cannot connect to Master (fatal for client)

#### Hybrid Mode
- Creates Node with Hybrid role (both Master and Client)
- Initializes both NodeManager and MasterConnector
- Can accept client connections AND connect to upstream Master
- Non-fatal if Master connection fails (continues as local Master)

### 4. Graceful Shutdown
Implemented proper shutdown hook registration with priorities:
1. HTTP Server (PriorityCritical)
2. MasterConnector (PriorityCritical) - if client/hybrid
3. NodeManager (PriorityHigh) - if master/hybrid
4. Node (PriorityHigh)
5. Models (PriorityHigh)
6. Processes (PriorityNormal)
7. Logger (PriorityLow)

### 5. Configuration Integration
- Reads cfg.Node.Role for role determination
- Uses cfg.Node.MasterRole for master-specific config (port, SSL, API key)
- Uses cfg.Node.ClientRole for client-specific config (master address, heartbeat)
- Auto-generates node ID if set to "auto"
- Builds NodeConfig from application configuration

## Key Design Decisions

### Backward Compatibility
- Existing mode command-line parameter still works
- Old config files without Node section use standalone mode
- No breaking changes to existing standalone deployments

### Error Handling
- Client mode: Master connection failure is fatal
- Hybrid mode: Master connection failure is warning only
- Standalone/Master: No external dependencies

### Component Lifecycle
1. Initialize: Create all components (no goroutines started)
2. Start: Begin all goroutines in correct order
3. Shutdown: Stop in reverse order of dependencies

## Files Modified
- cmd/shepherd/main.go - Main refactoring
- internal/server/server.go - Added RegisterMasterHandler() method

## Dependencies
- Uses existing internal/node package for Node management
- Uses existing internal/master package for NodeManager
- Uses existing internal/client package for MasterConnector
- Uses existing internal/shutdown package for graceful shutdown
