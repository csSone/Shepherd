// Package client provides client node functionality for connecting to Shepherd master nodes.
//
// Deprecated: This package is deprecated and will be removed in a future version.
// Please use the unified Node architecture in github.com/shepherd-project/shepherd/Shepherd/internal/node instead.
//
// Migration Guide:
//   - client.MasterConnector → Use node.Node with Client role
//   - client.CommandHandler → Use node.CommandExecutor
//   - Client connection is automatically managed by the Node
//
// The Node architecture provides automatic connection management, built-in heartbeat
// and command polling with unified configuration for all node roles.
package client
