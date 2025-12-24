#!/bin/bash
# generate_testdata.sh - Generate sample procfs/sysfs files for testing.
#
# Creates realistic mock data for all files the profiler reads.
# Output goes to testdata/ directory, mirroring real system paths.

set -euo pipefail

OUTDIR="${1:-testdata}"

echo "Generating test data in $OUTDIR..."

# Create directory structure
mkdir -p "$OUTDIR"/{proc,sys/fs/cgroup,sys/devices/system/cpu,sys/class/dmi/id}
mkdir -p "$OUTDIR"/proc/{1,42,1337,net,self}
mkdir -p "$OUTDIR"/sys/fs/cgroup/{cpuacct,memory,blkio}
mkdir -p "$OUTDIR"/sys/devices/system/cpu/{cpu0,cpu1,cpu2,cpu3}/{cache/index0,cache/index1,cache/index2,cache/index3,cpufreq}

# =============================================================================
# /proc/stat - CPU statistics
# =============================================================================
cat > "$OUTDIR/proc/stat" << 'EOF'
cpu  10132153 290696 3084719 46828483 16683 0 25195 0 0 0
cpu0 2563470 73024 771178 11706192 4162 0 6297 0 0 0
cpu1 2520923 72652 770215 11711610 4177 0 6285 0 0 0
cpu2 2523933 72424 771628 11705234 4170 0 6314 0 0 0
cpu3 2523827 72596 771698 11705447 4174 0 6299 0 0 0
intr 620438388 38 0 0 0 0 0 0 0 1 0 0 0 156 0 0 0 28 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 1234567890
btime 1703462400
processes 123456
procs_running 3
procs_blocked 0
softirq 159613647 0 58361955 16888 39619257 14796 0 4275573 33379166 0 23926012
EOF

# =============================================================================
# /proc/loadavg - Load averages
# =============================================================================
cat > "$OUTDIR/proc/loadavg" << 'EOF'
1.52 1.23 0.98 3/892 12345
EOF

# =============================================================================
# /proc/meminfo - Memory statistics
# =============================================================================
cat > "$OUTDIR/proc/meminfo" << 'EOF'
MemTotal:       32768000 kB
MemFree:         4096000 kB
MemAvailable:   16384000 kB
Buffers:         1024000 kB
Cached:          8192000 kB
SwapCached:       512000 kB
Active:         12288000 kB
Inactive:        8192000 kB
Active(anon):    8192000 kB
Inactive(anon):  2048000 kB
Active(file):    4096000 kB
Inactive(file):  6144000 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:       8388608 kB
SwapFree:        7340032 kB
Dirty:             12288 kB
Writeback:             0 kB
AnonPages:      10240000 kB
Mapped:          2048000 kB
Shmem:            512000 kB
KReclaimable:     768000 kB
Slab:            1024000 kB
SReclaimable:     768000 kB
SUnreclaim:       256000 kB
KernelStack:       16384 kB
PageTables:        65536 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:    24576000 kB
Committed_AS:   20480000 kB
VmallocTotal:   34359738367 kB
VmallocUsed:       65536 kB
VmallocChunk:          0 kB
Percpu:            32768 kB
HardwareCorrupted:     0 kB
AnonHugePages:   2097152 kB
ShmemHugePages:        0 kB
ShmemPmdMapped:        0 kB
FileHugePages:         0 kB
FilePmdMapped:         0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
Hugetlb:               0 kB
DirectMap4k:      524288 kB
DirectMap2M:    16777216 kB
DirectMap1G:    16777216 kB
EOF

# =============================================================================
# /proc/vmstat - Virtual memory statistics
# =============================================================================
cat > "$OUTDIR/proc/vmstat" << 'EOF'
nr_free_pages 1024000
nr_zone_inactive_anon 512000
nr_zone_active_anon 2048000
nr_zone_inactive_file 1536000
nr_zone_active_file 1024000
nr_zone_unevictable 0
nr_zone_write_pending 3072
nr_mlock 0
nr_bounce 0
nr_zspages 0
nr_free_cma 0
numa_hit 123456789
numa_miss 0
numa_foreign 0
numa_interleave 12345
numa_local 123456789
numa_other 0
nr_inactive_anon 512000
nr_active_anon 2048000
nr_inactive_file 1536000
nr_active_file 1024000
nr_unevictable 0
nr_slab_reclaimable 192000
nr_slab_unreclaimable 64000
nr_isolated_anon 0
nr_isolated_file 0
workingset_nodes 12345
workingset_refault_anon 67890
workingset_refault_file 234567
workingset_activate_anon 12345
workingset_activate_file 67890
workingset_restore_anon 1234
workingset_restore_file 5678
workingset_nodereclaim 0
nr_anon_pages 2560000
nr_mapped 512000
nr_file_pages 2816000
nr_dirty 3072
nr_writeback 0
nr_writeback_temp 0
nr_shmem 128000
nr_shmem_hugepages 0
nr_shmem_pmdmapped 0
nr_file_hugepages 0
nr_file_pmdmapped 0
nr_anon_transparent_hugepages 512
nr_vmscan_write 12345
nr_vmscan_immediate_reclaim 0
nr_dirtied 98765432
nr_written 87654321
nr_kernel_misc_reclaimable 0
nr_foll_pin_acquired 0
nr_foll_pin_released 0
nr_kernel_stack 4096
pgpgin 123456789
pgpgout 234567890
pswpin 12345
pswpout 23456
pgalloc_dma 0
pgalloc_dma32 12345678
pgalloc_normal 987654321
pgalloc_movable 0
allocstall_dma 0
allocstall_dma32 0
allocstall_normal 123
allocstall_movable 0
pgskip_dma 0
pgskip_dma32 0
pgskip_normal 0
pgskip_movable 0
pgfree 1234567890
pgactivate 12345678
pgdeactivate 2345678
pglazyfree 123456
pgfault 987654321
pgmajfault 12345
pglazyfreed 98765
pgrefill 1234567
pgreuse 23456789
pgsteal_kswapd 12345678
pgsteal_direct 234567
pgdemote_kswapd 0
pgdemote_direct 0
pgscan_kswapd 23456789
pgscan_direct 345678
pgscan_direct_throttle 0
pgscan_anon 1234567
pgscan_file 22222222
pgsteal_anon 123456
pgsteal_file 12222222
zone_reclaim_failed 0
pginodesteal 12345
slabs_scanned 0
kswapd_inodesteal 234567
kswapd_low_wmark_hit_quickly 1234
kswapd_high_wmark_hit_quickly 567
pageoutrun 12345
pgrotated 2345
drop_pagecache 12
drop_slab 34
oom_kill 0
numa_pte_updates 0
numa_huge_pte_updates 0
numa_hint_faults 0
numa_hint_faults_local 0
numa_pages_migrated 0
pgmigrate_success 12345
pgmigrate_fail 123
thp_migration_success 12
thp_migration_fail 0
thp_migration_split 0
compact_migrate_scanned 1234567
compact_free_scanned 2345678
compact_isolated 123456
compact_stall 123
compact_fail 12
compact_success 111
compact_daemon_wake 234
compact_daemon_migrate_scanned 345678
compact_daemon_free_scanned 456789
htlb_buddy_alloc_success 0
htlb_buddy_alloc_fail 0
unevictable_pgs_culled 12345
unevictable_pgs_scanned 0
unevictable_pgs_rescued 1234
unevictable_pgs_mlocked 23456
unevictable_pgs_munlocked 23456
unevictable_pgs_cleared 0
unevictable_pgs_stranded 0
thp_fault_alloc 12345
thp_fault_fallback 1234
thp_fault_fallback_charge 0
thp_collapse_alloc 2345
thp_collapse_alloc_failed 123
thp_file_alloc 0
thp_file_fallback 0
thp_file_fallback_charge 0
thp_file_mapped 0
thp_split_page 234
thp_split_page_failed 12
thp_deferred_split_page 1234
thp_split_pmd 345
thp_split_pud 0
thp_zero_page_alloc 123
thp_zero_page_alloc_failed 0
thp_swpout 12
thp_swpout_fallback 34
balloon_inflate 0
balloon_deflate 0
balloon_migrate 0
swap_ra 12345
swap_ra_hit 11111
EOF

# =============================================================================
# /proc/diskstats - Disk I/O statistics
# =============================================================================
cat > "$OUTDIR/proc/diskstats" << 'EOF'
   8       0 sda 1234567 89012 345678901 234567 2345678 901234 567890123 345678 0 456789 580245 0 0 0 0 12345 67890
   8       1 sda1 123456 7890 12345678 23456 234567 89012 34567890 34567 0 45678 58024 0 0 0 0 1234 6789
   8       2 sda2 1111111 81122 333333223 211111 2111111 812222 533322233 311111 0 411111 522221 0 0 0 0 11111 61111
 259       0 nvme0n1 2345678 123456 789012345 345678 3456789 234567 890123456 456789 0 567890 802467 0 0 0 0 23456 78901
 259       1 nvme0n1p1 234567 12345 67890123 34567 345678 23456 78901234 45678 0 56789 80246 0 0 0 0 2345 7890
 259       2 nvme0n1p2 2111111 111111 721122222 311111 3111111 211111 811222222 411111 0 511101 722221 0 0 0 0 21111 71011
   7       0 loop0 123 0 1234 12 0 0 0 0 0 12 12 0 0 0 0 0 0
   7       1 loop1 456 0 4567 45 0 0 0 0 0 45 45 0 0 0 0 0 0
EOF

# =============================================================================
# /proc/net/dev - Network interface statistics
# =============================================================================
cat > "$OUTDIR/proc/net/dev" << 'EOF'
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 12345678901  1234567    0    0    0     0          0         0 12345678901  1234567    0    0    0     0       0          0
  eth0: 98765432101 23456789   12   34    0     0          0      5678 87654321012 12345678   56   78    0     0       0          0
docker0: 1234567890  234567    0    0    0     0          0         0  2345678901  345678    0    0    0     0       0          0
veth1234: 123456789   12345    0    0    0     0          0         0   234567890   23456    0    0    0     0       0          0
EOF

# =============================================================================
# /proc/cpuinfo - CPU information
# =============================================================================
cat > "$OUTDIR/proc/cpuinfo" << 'EOF'
processor	: 0
vendor_id	: GenuineIntel
cpu family	: 6
model		: 106
model name	: Intel(R) Xeon(R) Platinum 8375C CPU @ 2.90GHz
stepping	: 6
microcode	: 0xd0003a5
cpu MHz		: 2900.000
cache size	: 55296 KB
physical id	: 0
siblings	: 4
core id		: 0
cpu cores	: 4
apicid		: 0
initial apicid	: 0
fpu		: yes
fpu_exception	: yes
cpuid level	: 27
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology nonstop_tsc cpuid aperfmperf tsc_known_freq pni pclmulqdq monitor ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single ssbd ibrs ibpb stibp ibrs_enhanced fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid avx512f avx512dq rdseed adx smap avx512ifma clflushopt clwb avx512cd sha_ni avx512bw avx512vl xsaveopt xsavec xgetbv1 xsaves wbnoinvd ida arat avx512vbmi pku ospke avx512_vbmi2 gfni vaes vpclmulqdq avx512_vnni avx512_bitalg tme avx512_vpopcntdq rdpid md_clear flush_l1d arch_capabilities
bugs		: spectre_v1 spectre_v2 spec_store_bypass swapgs
bogomips	: 5800.00
clflush size	: 64
cache_alignment	: 64
address sizes	: 46 bits physical, 48 bits virtual
power management:

processor	: 1
vendor_id	: GenuineIntel
cpu family	: 6
model		: 106
model name	: Intel(R) Xeon(R) Platinum 8375C CPU @ 2.90GHz
stepping	: 6
microcode	: 0xd0003a5
cpu MHz		: 2900.000
cache size	: 55296 KB
physical id	: 0
siblings	: 4
core id		: 1
cpu cores	: 4
apicid		: 2
initial apicid	: 2
fpu		: yes
fpu_exception	: yes
cpuid level	: 27
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology nonstop_tsc cpuid aperfmperf tsc_known_freq pni pclmulqdq monitor ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single ssbd ibrs ibpb stibp ibrs_enhanced fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid avx512f avx512dq rdseed adx smap avx512ifma clflushopt clwb avx512cd sha_ni avx512bw avx512vl xsaveopt xsavec xgetbv1 xsaves wbnoinvd ida arat avx512vbmi pku ospke avx512_vbmi2 gfni vaes vpclmulqdq avx512_vnni avx512_bitalg tme avx512_vpopcntdq rdpid md_clear flush_l1d arch_capabilities
bugs		: spectre_v1 spectre_v2 spec_store_bypass swapgs
bogomips	: 5800.00
clflush size	: 64
cache_alignment	: 64
address sizes	: 46 bits physical, 48 bits virtual
power management:

processor	: 2
vendor_id	: GenuineIntel
cpu family	: 6
model		: 106
model name	: Intel(R) Xeon(R) Platinum 8375C CPU @ 2.90GHz
stepping	: 6
microcode	: 0xd0003a5
cpu MHz		: 2900.000
cache size	: 55296 KB
physical id	: 0
siblings	: 4
core id		: 2
cpu cores	: 4
apicid		: 4
initial apicid	: 4
fpu		: yes
fpu_exception	: yes
cpuid level	: 27
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology nonstop_tsc cpuid aperfmperf tsc_known_freq pni pclmulqdq monitor ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single ssbd ibrs ibpb stibp ibrs_enhanced fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid avx512f avx512dq rdseed adx smap avx512ifma clflushopt clwb avx512cd sha_ni avx512bw avx512vl xsaveopt xsavec xgetbv1 xsaves wbnoinvd ida arat avx512vbmi pku ospke avx512_vbmi2 gfni vaes vpclmulqdq avx512_vnni avx512_bitalg tme avx512_vpopcntdq rdpid md_clear flush_l1d arch_capabilities
bugs		: spectre_v1 spectre_v2 spec_store_bypass swapgs
bogomips	: 5800.00
clflush size	: 64
cache_alignment	: 64
address sizes	: 46 bits physical, 48 bits virtual
power management:

processor	: 3
vendor_id	: GenuineIntel
cpu family	: 6
model		: 106
model name	: Intel(R) Xeon(R) Platinum 8375C CPU @ 2.90GHz
stepping	: 6
microcode	: 0xd0003a5
cpu MHz		: 2900.000
cache size	: 55296 KB
physical id	: 0
siblings	: 4
core id		: 3
cpu cores	: 4
apicid		: 6
initial apicid	: 6
fpu		: yes
fpu_exception	: yes
cpuid level	: 27
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon rep_good nopl xtopology nonstop_tsc cpuid aperfmperf tsc_known_freq pni pclmulqdq monitor ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single ssbd ibrs ibpb stibp ibrs_enhanced fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid avx512f avx512dq rdseed adx smap avx512ifma clflushopt clwb avx512cd sha_ni avx512bw avx512vl xsaveopt xsavec xgetbv1 xsaves wbnoinvd ida arat avx512vbmi pku ospke avx512_vbmi2 gfni vaes vpclmulqdq avx512_vnni avx512_bitalg tme avx512_vpopcntdq rdpid md_clear flush_l1d arch_capabilities
bugs		: spectre_v1 spectre_v2 spec_store_bypass swapgs
bogomips	: 5800.00
clflush size	: 64
cache_alignment	: 64
address sizes	: 46 bits physical, 48 bits virtual
power management:
EOF

# =============================================================================
# /proc/version - Kernel version
# =============================================================================
cat > "$OUTDIR/proc/version" << 'EOF'
Linux version 6.5.0-44-generic (buildd@lcy02-amd64-051) (x86_64-linux-gnu-gcc-12 (Ubuntu 12.3.0-1ubuntu1~23.04) 12.3.0, GNU ld (GNU Binutils for Ubuntu) 2.40) #44-Ubuntu SMP PREEMPT_DYNAMIC Fri Jun  7 15:10:09 UTC 2024
EOF

# =============================================================================
# /proc/self/cgroup - Container cgroup info (Docker example)
# =============================================================================
cat > "$OUTDIR/proc/self/cgroup" << 'EOF'
0::/docker/abc123def456789012345678901234567890123456789012345678901234
EOF

# =============================================================================
# Process 1 (init/systemd)
# =============================================================================
cat > "$OUTDIR/proc/1/stat" << 'EOF'
1 (systemd) S 0 1 1 0 -1 4194560 123456 78901234 123 456 7890 1234 56789 12345 20 0 1 0 12 234567890 12345 18446744073709551615 94123456789012 94123456901234 140123456789012 0 0 0 671173123 4096 1260 0 0 0 17 0 0 0 0 0 0 94123456912345 94123456923456 94123789012345 140123456789123 140123456789234 140123456789234 140123456789567 0
EOF

cat > "$OUTDIR/proc/1/status" << 'EOF'
Name:	systemd
Umask:	0000
State:	S (sleeping)
Tgid:	1
Ngid:	0
Pid:	1
PPid:	0
TracerPid:	0
Uid:	0	0	0	0
Gid:	0	0	0	0
FDSize:	256
Groups:	
NStgid:	1
NSpid:	1
NSpgid:	1
NSsid:	1
VmPeak:	  234568 kB
VmSize:	  234568 kB
VmLck:	       0 kB
VmPin:	       0 kB
VmHWM:	   12345 kB
VmRSS:	   12345 kB
RssAnon:	    5678 kB
RssFile:	    6667 kB
RssShmem:	       0 kB
VmData:	   18234 kB
VmStk:	     132 kB
VmExe:	    1416 kB
VmLib:	    9876 kB
VmPTE:	     128 kB
VmSwap:	       0 kB
HugetlbPages:	       0 kB
CoreDumping:	0
THP_enabled:	1
Threads:	1
SigQ:	0/123456
SigPnd:	0000000000000000
ShdPnd:	0000000000000000
SigBlk:	7be3c0fe28014a03
SigIgn:	0000000000001000
SigCgt:	00000001800004ec
CapInh:	0000000000000000
CapPrm:	000001ffffffffff
CapEff:	000001ffffffffff
CapBnd:	000001ffffffffff
CapAmb:	0000000000000000
NoNewPrivs:	0
Seccomp:	0
Seccomp_filters:	0
Speculation_Store_Bypass:	thread vulnerable
SpeculationIndirectBranch:	conditional enabled
Cpus_allowed:	f
Cpus_allowed_list:	0-3
Mems_allowed:	00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000001
Mems_allowed_list:	0
voluntary_ctxt_switches:	123456
nonvoluntary_ctxt_switches:	7890
EOF

cat > "$OUTDIR/proc/1/statm" << 'EOF'
58642 3086 1666 354 0 4558 0
EOF

printf 'systemd\x00--switched-root\x00--system\x00--deserialize\x0028\x00' > "$OUTDIR/proc/1/cmdline"

# =============================================================================
# Process 42 (example user process)
# =============================================================================
cat > "$OUTDIR/proc/42/stat" << 'EOF'
42 (python3) S 1 42 42 0 -1 4194304 567890 0 123 0 45678 12345 0 0 20 0 4 0 1234 1234567890 234567 18446744073709551615 94234567890123 94234567901234 140234567890123 0 0 0 0 0 65536 1 0 0 17 2 0 0 0 0 0 94234567912345 94234567923456 94234789012345 140234567890234 140234567890345 140234567890345 140234567890678 0
EOF

cat > "$OUTDIR/proc/42/status" << 'EOF'
Name:	python3
Umask:	0022
State:	S (sleeping)
Tgid:	42
Ngid:	0
Pid:	42
PPid:	1
TracerPid:	0
Uid:	1000	1000	1000	1000
Gid:	1000	1000	1000	1000
FDSize:	64
Groups:	4 24 27 30 46 100 118 1000
NStgid:	42
NSpid:	42
NSpgid:	42
NSsid:	42
VmPeak:	 1234568 kB
VmSize:	 1234568 kB
VmLck:	       0 kB
VmPin:	       0 kB
VmHWM:	  234567 kB
VmRSS:	  234567 kB
RssAnon:	  200000 kB
RssFile:	   34567 kB
RssShmem:	       0 kB
VmData:	  456789 kB
VmStk:	     136 kB
VmExe:	    2828 kB
VmLib:	   45678 kB
VmPTE:	    1024 kB
VmSwap:	       0 kB
HugetlbPages:	       0 kB
CoreDumping:	0
THP_enabled:	1
Threads:	4
SigQ:	1/123456
SigPnd:	0000000000000000
ShdPnd:	0000000000000000
SigBlk:	0000000000000000
SigIgn:	0000000001001000
SigCgt:	0000000180014a07
CapInh:	0000000000000000
CapPrm:	0000000000000000
CapEff:	0000000000000000
CapBnd:	000001ffffffffff
CapAmb:	0000000000000000
NoNewPrivs:	0
Seccomp:	0
Seccomp_filters:	0
Speculation_Store_Bypass:	thread vulnerable
SpeculationIndirectBranch:	conditional enabled
Cpus_allowed:	f
Cpus_allowed_list:	0-3
Mems_allowed:	00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000001
Mems_allowed_list:	0
voluntary_ctxt_switches:	98765
nonvoluntary_ctxt_switches:	4321
EOF

cat > "$OUTDIR/proc/42/statm" << 'EOF'
308642 58641 8641 707 0 114197 0
EOF

printf '/usr/bin/python3\x00-m\x00vllm.entrypoints.openai.api_server\x00--model\x00/app/model/\x00' > "$OUTDIR/proc/42/cmdline"

# =============================================================================
# Process 1337 (example GPU process)
# =============================================================================
cat > "$OUTDIR/proc/1337/stat" << 'EOF'
1337 (cuda_worker) R 42 42 42 0 -1 4194304 890123 0 456 0 78901 23456 0 0 20 0 1 0 5678 2345678901 345678 18446744073709551615 94345678901234 94345678912345 140345678901234 0 0 0 0 0 65536 1 0 0 17 1 0 0 0 0 0 94345678923456 94345678934567 94345890123456 140345678901345 140345678901456 140345678901456 140345678901789 0
EOF

cat > "$OUTDIR/proc/1337/status" << 'EOF'
Name:	cuda_worker
Umask:	0022
State:	R (running)
Tgid:	1337
Ngid:	0
Pid:	1337
PPid:	42
TracerPid:	0
Uid:	1000	1000	1000	1000
Gid:	1000	1000	1000	1000
FDSize:	128
Groups:	4 24 27 30 46 100 118 1000
NStgid:	1337
NSpid:	1337
NSpgid:	42
NSsid:	42
VmPeak:	 8765432 kB
VmSize:	 8765432 kB
VmLck:	       0 kB
VmPin:	       0 kB
VmHWM:	 4567890 kB
VmRSS:	 4567890 kB
RssAnon:	 4000000 kB
RssFile:	  567890 kB
RssShmem:	       0 kB
VmData:	 6789012 kB
VmStk:	     136 kB
VmExe:	    2828 kB
VmLib:	  123456 kB
VmPTE:	    8192 kB
VmSwap:	       0 kB
HugetlbPages:	       0 kB
CoreDumping:	0
THP_enabled:	1
Threads:	1
SigQ:	2/123456
SigPnd:	0000000000000000
ShdPnd:	0000000000000000
SigBlk:	0000000000000000
SigIgn:	0000000001001000
SigCgt:	0000000180014a07
CapInh:	0000000000000000
CapPrm:	0000000000000000
CapEff:	0000000000000000
CapBnd:	000001ffffffffff
CapAmb:	0000000000000000
NoNewPrivs:	0
Seccomp:	0
Seccomp_filters:	0
Speculation_Store_Bypass:	thread vulnerable
SpeculationIndirectBranch:	conditional enabled
Cpus_allowed:	f
Cpus_allowed_list:	0-3
Mems_allowed:	00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000001
Mems_allowed_list:	0
voluntary_ctxt_switches:	543210
nonvoluntary_ctxt_switches:	12345
EOF

cat > "$OUTDIR/proc/1337/statm" << 'EOF'
2191358 1141972 141972 707 0 1697253 0
EOF

printf 'cuda_worker\x00--batch-size\x0032\x00' > "$OUTDIR/proc/1337/cmdline"

# =============================================================================
# CPU sysfs - frequency and cache info
# =============================================================================
for cpu in 0 1 2 3; do
    echo "2900000" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cpufreq/scaling_cur_freq"
    
    # L1 data cache
    mkdir -p "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index0"
    echo "1" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index0/level"
    echo "Data" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index0/type"
    echo "48K" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index0/size"
    printf '%02x' "$cpu" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index0/shared_cpu_map"
    
    # L1 instruction cache
    mkdir -p "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index1"
    echo "1" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index1/level"
    echo "Instruction" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index1/type"
    echo "32K" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index1/size"
    printf '%02x' "$cpu" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index1/shared_cpu_map"
    
    # L2 cache (per-core)
    mkdir -p "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index2"
    echo "2" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index2/level"
    echo "Unified" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index2/type"
    echo "1280K" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index2/size"
    printf '%02x' "$cpu" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index2/shared_cpu_map"
    
    # L3 cache (shared across all cores)
    mkdir -p "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index3"
    echo "3" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index3/level"
    echo "Unified" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index3/type"
    echo "55296K" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index3/size"
    echo "0f" > "$OUTDIR/sys/devices/system/cpu/cpu$cpu/cache/index3/shared_cpu_map"
done

# =============================================================================
# DMI info
# =============================================================================
echo "EC2ABCD1-2345-6789-ABCD-EF0123456789" > "$OUTDIR/sys/class/dmi/id/product_uuid"

# =============================================================================
# Machine ID fallback
# =============================================================================
mkdir -p "$OUTDIR/etc"
echo "abc123def456789012345678901234ab" > "$OUTDIR/etc/machine-id"

# =============================================================================
# Cgroup v2 files
# =============================================================================
mkdir -p "$OUTDIR/sys/fs/cgroup"
echo "cpuset cpu io memory hugetlb pids rdma misc" > "$OUTDIR/sys/fs/cgroup/cgroup.controllers"

cat > "$OUTDIR/sys/fs/cgroup/cpu.stat" << 'EOF'
usage_usec 123456789012
user_usec 98765432109
system_usec 24691356903
nr_periods 0
nr_throttled 0
throttled_usec 0
nr_bursts 0
burst_usec 0
EOF

echo "8589934592" > "$OUTDIR/sys/fs/cgroup/memory.current"
echo "12884901888" > "$OUTDIR/sys/fs/cgroup/memory.peak"

cat > "$OUTDIR/sys/fs/cgroup/io.stat" << 'EOF'
8:0 rbytes=12345678901 wbytes=9876543210 rios=123456 wios=98765 dbytes=0 dios=0
259:0 rbytes=23456789012 wbytes=8765432109 rios=234567 wios=87654 dbytes=0 dios=0
EOF

# =============================================================================
# Cgroup v1 files (alternative)
# =============================================================================
mkdir -p "$OUTDIR/sys/fs/cgroup_v1"/{cpuacct,memory,blkio}

echo "1234567890123456789" > "$OUTDIR/sys/fs/cgroup_v1/cpuacct/cpuacct.usage"

cat > "$OUTDIR/sys/fs/cgroup_v1/cpuacct/cpuacct.stat" << 'EOF'
user 987654321
system 246913569
EOF

echo "308642000000 308641000000 308640000000 308639000000" > "$OUTDIR/sys/fs/cgroup_v1/cpuacct/cpuacct.usage_percpu"

echo "8589934592" > "$OUTDIR/sys/fs/cgroup_v1/memory/memory.usage_in_bytes"
echo "12884901888" > "$OUTDIR/sys/fs/cgroup_v1/memory/memory.max_usage_in_bytes"

cat > "$OUTDIR/sys/fs/cgroup_v1/blkio/blkio.throttle.io_service_bytes" << 'EOF'
8:0 Read 12345678901
8:0 Write 9876543210
8:0 Sync 11111111111
8:0 Async 11111111000
8:0 Discard 0
8:0 Total 22222222111
259:0 Read 23456789012
259:0 Write 8765432109
259:0 Sync 16111111061
259:0 Async 16111110060
259:0 Discard 0
259:0 Total 32222221121
Total 54444443232
EOF

# =============================================================================
# vLLM Prometheus metrics sample
# =============================================================================
cat > "$OUTDIR/vllm_metrics.txt" << 'EOF'
# HELP vllm:num_requests_running Number of requests currently running on GPU.
# TYPE vllm:num_requests_running gauge
vllm:num_requests_running{model_name="llama-3.2-1b"} 3
# HELP vllm:num_requests_waiting Number of requests waiting to be processed.
# TYPE vllm:num_requests_waiting gauge
vllm:num_requests_waiting{model_name="llama-3.2-1b"} 12
# HELP vllm:engine_sleep_state Sleep state of the engine.
# TYPE vllm:engine_sleep_state gauge
vllm:engine_sleep_state{model_name="llama-3.2-1b"} 1
# HELP vllm:kv_cache_usage_perc GPU KV-cache usage. 1 means 100 percent usage.
# TYPE vllm:kv_cache_usage_perc gauge
vllm:kv_cache_usage_perc{model_name="llama-3.2-1b"} 0.42
# HELP vllm:num_preemptions Total preemptions.
# TYPE vllm:num_preemptions counter
vllm:num_preemptions{model_name="llama-3.2-1b"} 5
# HELP vllm:prefix_cache_hits Prefix cache hits.
# TYPE vllm:prefix_cache_hits counter
vllm:prefix_cache_hits{model_name="llama-3.2-1b"} 1234
# HELP vllm:prefix_cache_queries Prefix cache queries.
# TYPE vllm:prefix_cache_queries counter
vllm:prefix_cache_queries{model_name="llama-3.2-1b"} 5678
# HELP vllm:request_success Successful requests.
# TYPE vllm:request_success counter
vllm:request_success{model_name="llama-3.2-1b"} 9876
# HELP vllm:prompt_tokens Total prompt tokens processed.
# TYPE vllm:prompt_tokens counter
vllm:prompt_tokens{model_name="llama-3.2-1b"} 1234567
# HELP vllm:generation_tokens Total generation tokens produced.
# TYPE vllm:generation_tokens counter
vllm:generation_tokens{model_name="llama-3.2-1b"} 2345678
# HELP vllm:time_to_first_token_seconds Histogram of time to first token in seconds.
# TYPE vllm:time_to_first_token_seconds histogram
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.001"} 0
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.005"} 12
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.01"} 45
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.025"} 234
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.05"} 567
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.075"} 789
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.1"} 890
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.15"} 901
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.2"} 912
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.3"} 923
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.4"} 934
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.5"} 945
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="0.75"} 956
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="1.0"} 967
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="2.5"} 978
vllm:time_to_first_token_seconds_bucket{model_name="llama-3.2-1b",le="+Inf"} 987
vllm:time_to_first_token_seconds_count{model_name="llama-3.2-1b"} 987
vllm:time_to_first_token_seconds_sum{model_name="llama-3.2-1b"} 123.456
# HELP vllm:e2e_request_latency_seconds Histogram of end-to-end request latency in seconds.
# TYPE vllm:e2e_request_latency_seconds histogram
vllm:e2e_request_latency_seconds_bucket{model_name="llama-3.2-1b",le="0.1"} 100
vllm:e2e_request_latency_seconds_bucket{model_name="llama-3.2-1b",le="0.5"} 500
vllm:e2e_request_latency_seconds_bucket{model_name="llama-3.2-1b",le="1.0"} 800
vllm:e2e_request_latency_seconds_bucket{model_name="llama-3.2-1b",le="2.5"} 950
vllm:e2e_request_latency_seconds_bucket{model_name="llama-3.2-1b",le="5.0"} 980
vllm:e2e_request_latency_seconds_bucket{model_name="llama-3.2-1b",le="10.0"} 990
vllm:e2e_request_latency_seconds_bucket{model_name="llama-3.2-1b",le="+Inf"} 1000
vllm:e2e_request_latency_seconds_count{model_name="llama-3.2-1b"} 1000
vllm:e2e_request_latency_seconds_sum{model_name="llama-3.2-1b"} 987.654
# HELP vllm:cache_config_info Cache configuration info
# TYPE vllm:cache_config_info gauge
vllm:cache_config_info{block_size="16",cache_dtype="auto",num_gpu_blocks="2048",num_cpu_blocks="512"} 1
EOF

echo ""
echo "Test data generated in $OUTDIR/"
echo ""
echo "Directory structure:"
find "$OUTDIR" -type f | head -40
echo "..."
echo ""
echo "To use in tests, set environment variables or mock file paths to point to $OUTDIR"
