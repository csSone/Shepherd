// Package master provides master node management for distributed Shepherd deployments.
//
// Deprecated: This package is deprecated and will be removed in a future version.
// Please use the unified Node architecture in github.com/shepherd-project/shepherd/Shepherd/internal/node instead.
//
// Migration Guide:
//   - master.NodeManager → Use node.Node with client registry
//   - master.HeartbeatManager → Use node.HeartbeatManager
//   - master.Scheduler → Use node.Subsystems
//   - master.Handler → Use api.NodeAdapter for API compatibility
//
// The Node architecture provides unified management for all node roles (standalone, master, client, hybrid)
// with simplified configuration and better scalability.
package master
