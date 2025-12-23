import glob
import os
from inference_profiler.collectors.base import BaseColletor


class ProcCollector(BaseColletor):
    @staticmethod
    def collect():
        process_list = []
        for pid_path in glob.glob('/proc/[0-9]*'):
            try:
                pid = int(os.path.basename(pid_path))
                stat_data = BaseColletor._read_file(os.path.join(pid_path, 'stat'))
                if not stat_data: continue

                rparen_index = stat_data.rfind(')')
                if rparen_index == -1: continue

                stat_parts = stat_data[rparen_index + 2:].split()
                if len(stat_parts) < 40: continue

                cmdline = BaseColletor._read_file(os.path.join(pid_path, 'cmdline')).replace('\x00', ' ').strip()
                status_data = BaseColletor._read_status(os.path.join(pid_path, 'status'))

                name = status_data.get('Name', stat_data[stat_data.find('(') + 1:rparen_index])

                process_list.append({
                    "pId": pid,
                    "pCmdline": cmdline,
                    "pName": name,
                    "pNumThreads": int(stat_parts[17]),
                    "pCpuTimeUserMode": int(stat_parts[11]) * BaseColletor.JIFFIES_PER_SECOND,
                    "pCpuTimeKernelMode": int(stat_parts[12]) * BaseColletor.JIFFIES_PER_SECOND,
                    "pChildrenUserMode": int(stat_parts[13]) * BaseColletor.JIFFIES_PER_SECOND,
                    "pChildrenKernelMode": int(stat_parts[14]) * BaseColletor.JIFFIES_PER_SECOND,
                    "pVoluntaryContextSwitches": status_data.get('voluntary_ctxt_switches', 0),
                    "pInvoluntaryContextSwitches": status_data.get('nonvoluntary_ctxt_switches', 0),
                    "pBlockIODelays": int(stat_parts[39]) * BaseColletor.JIFFIES_PER_SECOND,
                    "pVirtualMemoryBytes": int(stat_parts[20])
                })
            except Exception:
                continue
        return process_list