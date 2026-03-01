// Package fs implements filesystem snapshot handling for the DCC Protocol.
//
// The FilesystemSnapshot provides read-only snapshot layers for the capsule
// and tracks ephemeral /tmp files.
//
// Key features:
//   - Read-only snapshot from fs_snapshot_hash
//   - Includes: dependencies, compiled artifacts, static assets
//   - /tmp is ephemeral, tracked, hashed into OutputHash
//   - Replay reconstructs from hash
//
// Usage:
//
//	snapshot := fs.New("base-snapshot-hash")
//	
//	// Add files to snapshot
//	snapshot.AddFile("/app/main.py", []byte("code"))
//	snapshot.AddDirectory("/app/data")
//	
//	// Write ephemeral /tmp files
//	snapshot.WriteTmp("temp.txt", []byte("data"))
//	
//	// Get snapshot hash
//	fsHash := snapshot.Hash()
//	
//	// Get ephemeral hash (for OutputHash)
//	ephemeralHash := snapshot.EphemeralHash()
package fs
