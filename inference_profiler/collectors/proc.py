import glob
import json
import os

from .base import BaseCollector


class ProcCollector(BaseCollector):
    @staticmethod
    def collect():
        process_list = []
        process_count = 0

        for pid_path in glob.glob('/proc/[0-9]*'):
            try:
                pid = int(os.path.basename(pid_path))
                stat_data, t_stat_data = BaseCollector._probe_file(os.path.join(pid_path, 'stat'))
                cmdline, t_cmdline = BaseCollector._probe_file(os.path.join(pid_path, 'cmdline'))
                status_data, t_status_data = BaseCollector._parse_proc_kv(os.path.join(pid_path, 'status'))
                statm_data, t_statm = BaseCollector._probe_file(os.path.join(pid_path, 'statm'))

                if not stat_data:
                    continue

                rparen_index = stat_data.rfind(')')

                if rparen_index == -1:
                    continue

                stat_parts = stat_data[rparen_index + 2:].split()

                if len(stat_parts) < 40:
                    continue

                process_count += 1

                # Parse statm for resident set size
                # statm fields: size resident shared text lib data dt (all in pages)
                rss_pages = 0
                if statm_data:
                    statm_parts = statm_data.split()
                    if len(statm_parts) >= 2:
                        try:
                            rss_pages = int(statm_parts[1])
                        except ValueError:
                            pass

                # Get page size (usually 4096)
                page_size = os.sysconf('SC_PAGE_SIZE') if hasattr(os, 'sysconf') else 4096
                rss_bytes = rss_pages * page_size

                process_list.append({
                    # Process identification
                    "pId": pid,
                    "pName": status_data.get('Name', stat_data[stat_data.find('(') + 1:rparen_index]),
                    "pCmdline": cmdline.replace('\x00', ' ').strip() if cmdline else "",
                    "pNumThreads": int(stat_parts[17]),
                    "tpNumThreads": t_stat_data,
                    "pCpuTimeUserMode": int(stat_parts[11]) * BaseCollector.JIFFIES_PER_SECOND,
                    "tpCpuTimeUserMode": t_stat_data,
                    "pCpuTimeKernelMode": int(stat_parts[12]) * BaseCollector.JIFFIES_PER_SECOND,
                    "tpCpuTimeKernelMode": t_stat_data,
                    "pChildrenUserMode": int(stat_parts[13]) * BaseCollector.JIFFIES_PER_SECOND,
                    "tpChildrenUserMode": t_stat_data,
                    "pChildrenKernelMode": int(stat_parts[14]) * BaseCollector.JIFFIES_PER_SECOND,
                    "tpChildrenKernelMode": t_stat_data,
                    "pVoluntaryContextSwitches": int(status_data.get('voluntary_ctxt_switches', 0)),
                    "tpVoluntaryContextSwitches": t_status_data,
                    "pNonvoluntaryContextSwitches": int(status_data.get('nonvoluntary_ctxt_switches', 0)),
                    "tpNonvoluntaryContextSwitches": t_status_data,
                    "pBlockIODelays": int(stat_parts[39]) * BaseCollector.JIFFIES_PER_SECOND,
                    "tpBlockIODelays": t_stat_data,
                    "pVirtualMemoryBytes": int(stat_parts[20]),
                    "tpVirtualMemoryBytes": t_stat_data,
                    "pResidentSetSize": rss_bytes,
                    "tpResidentSetSize": t_statm,
                })
            except Exception:
                BaseCollector.logger.exception("Failed to collect proc data")
                continue

        return {
            "pNumProcesses": process_count,
            "processes": [json.dumps(x).replace('"', '\"') for x in process_list]
        }