package memory

import (
	"fmt"
)

// AcquireDrawer attempts to lock a drawer for exclusive use by a unit.
// It creates a snapshot of the drawer state immediately.
func (c *Cabinet) AcquireDrawer(unitID int) (*DrawerLease, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find available drawer (not leased)
	// Priority: 1. Existing unleased in RAM.
	// Future: 2. Swap-in. 3. Create new.

	for _, drawerID := range c.RAMSlots {
		if drawerID == -1 {
			continue
		}
		drawer := &c.Drawers[drawerID]

		if drawer.Lease == nil {
			// Found candidate
			snapID, err := c.SnapshotDrawer(drawerID)
			if err != nil {
				return nil, err
			}

			lease := &DrawerLease{
				DrawerID:   drawerID,
				UnitID:     unitID,
				SnapshotID: snapID,
				IsActive:   true,
			}
			drawer.Lease = lease
			return lease, nil
		}
	}

	return nil, fmt.Errorf("no available drawers in RAM")
}

// CommitDrawer releases the lease and treats current state as final.
func (c *Cabinet) CommitDrawer(lease *DrawerLease) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !lease.IsActive {
		return fmt.Errorf("lease is inactive")
	}

	drawer := &c.Drawers[lease.DrawerID]
	if drawer.Lease != lease {
		return fmt.Errorf("lease mismatch for drawer %d", lease.DrawerID)
	}

	// Release lease (finalize)
	drawer.Lease = nil
	lease.IsActive = false

	// Free the snapshot as it is no longer needed for rollback
	delete(c.Snapshots, lease.SnapshotID)

	return nil
}

// RollbackDrawer reverts the drawer to the state at acquisition.
func (c *Cabinet) RollbackDrawer(lease *DrawerLease) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !lease.IsActive {
		return fmt.Errorf("lease is inactive")
	}

	drawer := &c.Drawers[lease.DrawerID]
	if drawer.Lease != lease {
		return fmt.Errorf("lease mismatch for drawer %d", lease.DrawerID)
	}

	// Restore from Snapshot
	err := c.RestoreDrawer(lease.DrawerID, lease.SnapshotID)
	if err != nil {
		return err
	}

	// Release lease after rollback
	drawer.Lease = nil
	lease.IsActive = false
	delete(c.Snapshots, lease.SnapshotID)

	return nil
}
