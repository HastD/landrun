package sandbox

import (
	"fmt"

	"github.com/landlock-lsm/go-landlock/landlock"
	"github.com/landlock-lsm/go-landlock/landlock/syscall"
	"github.com/zouuup/landrun/internal/log"
)

type Config struct {
	ReadOnlyPaths            []string
	ReadWritePaths           []string
	ReadOnlyExecutablePaths  []string
	ReadWriteExecutablePaths []string
	BindTCPPorts             []int
	ConnectTCPPorts          []int
	BestEffort               bool
}

// getReadWriteExecutableRights returns a full set of permissions including execution
func getReadWriteExecutableRights() landlock.AccessFSSet {
	accessRights := landlock.AccessFSSet(0)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSExecute)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSWriteFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSRemoveDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSRemoveFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeChar)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeReg)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeSock)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeFifo)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeBlock)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeSym)
	return accessRights
}

func getReadOnlyExecutableRights() landlock.AccessFSSet {
	accessRights := landlock.AccessFSSet(0)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSExecute)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadDir)
	return accessRights
}

// getReadOnlyRights returns permissions for read-only access
func getReadOnlyRights() landlock.AccessFSSet {
	accessRights := landlock.AccessFSSet(0)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadDir)
	return accessRights
}

// getReadWriteRights returns permissions for read-write access
func getReadWriteRights() landlock.AccessFSSet {
	accessRights := landlock.AccessFSSet(0)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSWriteFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSRemoveDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSRemoveFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeChar)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeReg)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeSock)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeFifo)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeBlock)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeSym)
	return accessRights
}

func Apply(cfg Config) error {
	log.Info("Sandbox config: %+v", cfg)

	// Get the most advanced Landlock version available
	llCfg := landlock.V5
	if cfg.BestEffort {
		llCfg = llCfg.BestEffort()
	}

	// Collect our rules
	var rules []landlock.Rule

	// Process executable paths
	for _, path := range cfg.ReadOnlyExecutablePaths {
		log.Debug("Adding read-only executable path: %s", path)
		rules = append(rules, landlock.PathAccess(getReadOnlyExecutableRights(), path))
	}

	for _, path := range cfg.ReadWriteExecutablePaths {
		log.Debug("Adding read-write executable path: %s", path)
		rules = append(rules, landlock.PathAccess(getReadWriteExecutableRights(), path))
	}

	// Process read-only paths
	for _, path := range cfg.ReadOnlyPaths {
		log.Debug("Adding read-only path: %s", path)
		rules = append(rules, landlock.PathAccess(getReadOnlyRights(), path))
	}

	// Process read-write paths
	for _, path := range cfg.ReadWritePaths {
		log.Debug("Adding read-write path: %s", path)
		rules = append(rules, landlock.PathAccess(getReadWriteRights(), path))
	}

	// Add rules for TCP port binding
	for _, port := range cfg.BindTCPPorts {
		log.Debug("Adding TCP bind port: %d", port)
		rules = append(rules, landlock.BindTCP(uint16(port)))
	}

	// Add rules for TCP connections
	for _, port := range cfg.ConnectTCPPorts {
		log.Debug("Adding TCP connect port: %d", port)
		rules = append(rules, landlock.ConnectTCP(uint16(port)))
	}

	// If we have no rules, just return
	if len(rules) == 0 {
		log.Error("No rules provided, applying default restrictive rules, this will restrict anything landlock can do.")
		err := llCfg.Restrict()
		if err != nil {
			return fmt.Errorf("failed to apply default Landlock restrictions: %w", err)
		}
		log.Info("Default restrictive Landlock rules applied successfully")
		return nil
	}

	// Apply all rules at once
	log.Debug("Applying Landlock restrictions")
	err := llCfg.Restrict(rules...)
	if err != nil {
		return fmt.Errorf("failed to apply Landlock restrictions: %w", err)
	}

	log.Info("Landlock restrictions applied successfully")
	return nil
}

// pathInSlice checks if a path exists in a slice of paths
func pathInSlice(path string, paths []string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}
