package services

import (
	"fmt"
	"strings"

	"incident-teller/internal/domain"
)

// FixRecommender provides structured, actionable remediation guidance
type FixRecommender struct {
	// Knowledge base of fixes per resource type
	immediateActions  map[domain.ResourceType][]string
	shortTermActions  map[domain.ResourceType][]string
	longTermActions   map[domain.ResourceType][]string
}

// NewFixRecommender creates a new fix recommender with built-in playbooks
func NewFixRecommender() *FixRecommender {
	fr := &FixRecommender{
		immediateActions: make(map[domain.ResourceType][]string),
		shortTermActions: make(map[domain.ResourceType][]string),
		longTermActions:  make(map[domain.ResourceType][]string),
	}
	
	fr.loadPlaybooks()
	return fr
}

// loadPlaybooks initializes the fix playbook database
func (fr *FixRecommender) loadPlaybooks() {
	// MEMORY playbooks
	fr.immediateActions[domain.ResourceMemory] = []string{
		"Identify top memory consumer: `ps aux --sort=-%mem | head -5`",
		"Check OOM killer logs: `dmesg | grep -i 'killed process'`",
		"If safe, restart highest memory process: `systemctl restart <service>`",
		"Clear page cache if needed: `echo 1 > /proc/sys/vm/drop_caches` (careful!)",
	}
	fr.shortTermActions[domain.ResourceMemory] = []string{
		"Analyze memory trends: `vmstat 1 10` and `free -h`",
		"Profile application heap usage with tools like pprof or valgrind",
		"Review recent deployments - rollback if memory leak introduced",
		"Adjust container memory limits if using Docker/K8s",
		"Enable swap if not already active: `swapon -a`",
	}
	fr.longTermActions[domain.ResourceMemory] = []string{
		"Implement memory limits for all services",
		"Set up proactive alerts at 70% and 85% memory usage",
		"Establish regular memory profiling in CI/CD pipeline",
		"Plan horizontal scaling for memory-intensive services",
		"Document baseline memory usage per service",
	}

	// DISK playbooks
	fr.immediateActions[domain.ResourceDisk] = []string{
		"Find largest files NOW: `du -ahx / | sort -rh | head -20`",
		"Clear Docker logs: `docker system prune -a --volumes` (if using Docker)",
		"Truncate largest log file: `truncate -s 0 /path/to/large.log`",
		"Delete old journal logs: `journalctl --vacuum-size=500M`",
	}
	fr.shortTermActions[domain.ResourceDisk] = []string{
		"Set up log rotation for all services: edit `/etc/logrotate.d/`",
		"Identify space hogs: `ncdu /` (interactive disk usage)",
		"Move logs to separate partition or remote log aggregator",
		"Archive old data to object storage (S3, GCS, etc.)",
		"Expand disk volume if on cloud infrastructure",
	}
	fr.longTermActions[domain.ResourceDisk] = []string{
		"Implement centralized logging (Loki, ELK, etc.)",
		"Set up disk space alerts at 75% and 90%",
		"Automate log cleanup with cron jobs",
		"Use volume quotas for multi-tenant systems",
		"Plan disk capacity based on growth projections",
	}

	// CPU playbooks
	fr.immediateActions[domain.ResourceCPU] = []string{
		"Identify CPU hog: `top -o %CPU` or `htop`",
		"Check for runaway processes: `ps aux | awk '{if($3>80) print $0}'`",
		"Nice down non-critical processes: `renice +10 -p <PID>`",
		"Kill runaway process if confirmed safe: `kill -TERM <PID>`",
	}
	fr.shortTermActions[domain.ResourceCPU] = []string{
		"Profile CPU usage: `perf top` or application profiler",
		"Review cron jobs: `crontab -l` and `/etc/cron.d/*`",
		"Check for cryptominers: `ps aux | grep -E 'xmrig|minergate'`",
		"Optimize database queries causing high CPU",
		"Consider enabling CPU throttling for background tasks",
	}
	fr.longTermActions[domain.ResourceCPU] = []string{
		"Implement auto-scaling based on CPU metrics",
		"Set CPU limits for all containerized services",
		"Optimize application code hot paths (profiling-guided)",
		"Use CPU affinity for latency-sensitive processes",
		"Document baseline CPU usage patterns",
	}

	// NETWORK playbooks
	fr.immediateActions[domain.ResourceNetwork] = []string{
		"Check interface status: `ip -s link show`",
		"Identify top bandwidth users: `iftop` or `nethogs`",
		"Block suspicious IPs: `iptables -A INPUT -s <IP> -j DROP`",
		"Check for network errors: `netstat -i` (look for dropped packets)",
	}
	fr.shortTermActions[domain.ResourceNetwork] = []string{
		"Analyze traffic patterns: `tcpdump -i any -c 1000 -w capture.pcap`",
		"Review firewall rules: `iptables -L -n -v`",
		"Check DNS resolution: `dig @8.8.8.8 <domain>` and `/etc/resolv.conf`",
		"Test connectivity to dependencies: `nc -zv <host> <port>`",
		"Review recent network configuration changes",
	}
	fr.longTermActions[domain.ResourceNetwork] = []string{
		"Implement network monitoring (Prometheus node_exporter)",
		"Set up DDoS protection (CloudFlare, AWS Shield, etc.)",
		"Use connection pooling in applications",
		"Implement rate limiting on public endpoints",
		"Document network topology and dependencies",
	}

	// PROCESS playbooks
	fr.immediateActions[domain.ResourceProcess] = []string{
		"Restart failed service: `systemctl restart <service>`",
		"Check service status: `systemctl status <service>`",
		"View recent logs: `journalctl -u <service> -n 50 --no-pager`",
		"Verify process is running: `pgrep -a <process>`",
	}
	fr.shortTermActions[domain.ResourceProcess] = []string{
		"Review service configuration files in `/etc/<service>/`",
		"Check process limits: `cat /proc/<PID>/limits`",
		"Increase file descriptors if needed: edit `/etc/security/limits.conf`",
		"Review recent deployments - rollback if unstable",
		"Enable core dumps for crash analysis: `ulimit -c unlimited`",
	}
	fr.longTermActions[domain.ResourceProcess] = []string{
		"Implement process monitoring with supervisor/systemd",
		"Set up automatic restart on failure",
		"Use health checks and readiness probes",
		"Implement circuit breakers for dependency failures",
		"Document service dependencies and startup order",
	}
}

// RecommendFixes generates actionable fixes based on root cause and blast radius
func (fr *FixRecommender) RecommendFixes(
	rootCause RootCauseCandidate,
	blastRadius EnhancedBlastRadiusAnalysis,
) ActionableFix {
	
	resourceType := rootCause.Alert.ResourceType
	
	// Get base playbook
	immediate := fr.immediateActions[resourceType]
	shortTerm := fr.shortTermActions[resourceType]
	longTerm := fr.longTermActions[resourceType]

	// Enhance based on blast radius
	immediate, shortTerm = fr.enhanceForCascade(
		immediate, shortTerm, blastRadius, resourceType,
	)

	// Add context-specific actions
	immediate = fr.addContextualActions(immediate, rootCause)

	// Determine complexity
	complexity := fr.determineComplexity(blastRadius)
	
	// Estimate time to resolve
	estimatedTime := fr.estimateResolutionTime(blastRadius, complexity)

	return ActionableFix{
		ImmediateFix:           immediate,
		ShortTermFix:           shortTerm,
		LongTermFix:            longTerm,
		RootCauseType:          resourceType,
		FixComplexity:          complexity,
		EstimatedTimeToResolve: estimatedTime,
	}
}

// enhanceForCascade adds cascade-specific mitigation steps
func (fr *FixRecommender) enhanceForCascade(
	immediate, shortTerm []string,
	blastRadius EnhancedBlastRadiusAnalysis,
	rootCauseType domain.ResourceType,
) ([]string, []string) {
	
	if blastRadius.CascadeDepth == 0 {
		return immediate, shortTerm
	}

	// Add cascade mitigation to immediate actions
	cascadeActions := []string{
		fmt.Sprintf("⚠️ CASCADE DETECTED (%d levels) - prioritize root cause", blastRadius.CascadeDepth),
	}
	
	// Add resource-specific cascade mitigation
	for _, affected := range blastRadius.IndirectlyAffected {
		if strings.Contains(affected.Name, "CPU") {
			cascadeActions = append(cascadeActions, 
				"Monitor CPU recovery after root cause fix")
		}
		if strings.Contains(affected.Name, "MEMORY") {
			cascadeActions = append(cascadeActions, 
				"Watch for memory stabilization - may need manual restart")
		}
	}
	
	immediate = append(cascadeActions, immediate...)
	
	// Add monitoring to short-term
	shortTerm = append(shortTerm,
		fmt.Sprintf("Monitor all %d affected resources for recovery", 
			len(blastRadius.DirectlyAffected) + len(blastRadius.IndirectlyAffected)),
	)

	return immediate, shortTerm
}

// addContextualActions adds alert-specific actions
func (fr *FixRecommender) addContextualActions(
	actions []string,
	rootCause RootCauseCandidate,
) []string {
	
	alert := rootCause.Alert
	
	// Add host-specific context
	if alert.Host != "" {
		actions = append([]string{
			fmt.Sprintf("🎯 Target host: %s", alert.Host),
		}, actions...)
	}

	// Add value-based urgency
	if alert.Value >= 95.0 {
		actions = append([]string{
			fmt.Sprintf("🚨 CRITICAL: %s at %.1f%% - IMMEDIATE action required", 
				alert.ResourceType, alert.Value),
		}, actions...)
	} else if alert.Value >= 85.0 {
		actions = append([]string{
			fmt.Sprintf("⚠️ HIGH: %s at %.1f%% - act within 5 minutes", 
				alert.ResourceType, alert.Value),
		}, actions...)
	}

	return actions
}

// determineComplexity assesses fix complexity
func (fr *FixRecommender) determineComplexity(blastRadius EnhancedBlastRadiusAnalysis) string {
	score := 0

	// Multiple hosts = more complex
	if len(blastRadius.AffectedHosts) > 1 {
		score += 2
	}

	// Cascade increases complexity
	score += blastRadius.CascadeDepth

	// Critical alerts increase complexity
	if blastRadius.CriticalAlerts > 2 {
		score += 2
	}

	// High impact score
	if blastRadius.ImpactScore > 75 {
		score += 2
	}

	switch {
	case score <= 2:
		return "Simple (single resource, localized)"
	case score <= 5:
		return "Moderate (multiple resources or cascade)"
	default:
		return "Complex (widespread impact, deep cascade)"
	}
}

// estimateResolutionTime predicts time to full resolution
func (fr *FixRecommender) estimateResolutionTime(
	blastRadius EnhancedBlastRadiusAnalysis,
	complexity string,
) string {
	
	switch complexity {
	case "Simple (single resource, localized)":
		return "5-15 minutes (if playbook followed)"
	case "Moderate (multiple resources or cascade)":
		return "15-45 minutes (requires coordination)"
	default:
		return "45+ minutes (major incident, multiple interventions needed)"
	}
}

// FormatActionableFix creates formatted output for SREs
func FormatActionableFix(fix ActionableFix) string {
	var output strings.Builder

	output.WriteString("╔═══════════════════════════════════════════════════════════════╗\n")
	output.WriteString("║                 ACTIONABLE FIX PLAYBOOK                       ║\n")
	output.WriteString("╚═══════════════════════════════════════════════════════════════╝\n\n")

	output.WriteString(fmt.Sprintf("Root Cause: %s\n", fix.RootCauseType))
	output.WriteString(fmt.Sprintf("Complexity: %s\n", fix.FixComplexity))
	output.WriteString(fmt.Sprintf("Est. Time to Resolve: %s\n\n", fix.EstimatedTimeToResolve))

	output.WriteString("🔴 IMMEDIATE FIX (NOW - within 5 minutes)\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	for i, action := range fix.ImmediateFix {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, action))
	}
	output.WriteString("\n")

	output.WriteString("🟡 SHORT-TERM FIX (TODAY - within 8 hours)\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	for i, action := range fix.ShortTermFix {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, action))
	}
	output.WriteString("\n")

	output.WriteString("🟢 LONG-TERM PREVENTION (ONGOING)\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	for i, action := range fix.LongTermFix {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, action))
	}
	output.WriteString("\n")

	return output.String()
}
