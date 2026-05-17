#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "pandas",
#     "pyarrow",
# ]
# ///
"""
Collect InferenceProfiler JSONL + results → parquet files.

Outputs: runs.raw.parquet, metrics.raw.parquet,
         requests.raw.parquet, tokens.raw.parquet, procs.parquet

Usage: ./collect.py [--out DIR] DIR [DIR ...]
"""

import json, sys
import pandas as pd
from pathlib import Path


def flatten(obj, prefix=""):
    flat, stack = {}, [(prefix, obj)]
    while stack:
        pfx, d = stack.pop()
        for k, v in d.items():
            fk = f"{pfx}{k[0].upper()}{k[1:]}" if pfx else k
            if isinstance(v, dict):
                stack.append((fk, v))
            elif isinstance(v, list):
                for i, el in enumerate(v):
                    if isinstance(el, dict):
                        stack.append((f"{fk}{i}", el))
                    else:
                        flat[f"{fk}{i}"] = el
            else:
                flat[fk] = v
    return flat


def strip_v(df):
    cols = set(df.columns)
    rn = {c: c[:-1] for c in cols if c.endswith("V") and c[:-1] + "T" in cols}
    return df.rename(columns=rn) if rn else df


def gpu_procs(nvidia):
    by_pid = {}
    for dev in nvidia or []:
        pb = dev.get("Processes") or {}
        for gp in pb.get("List") or []:
            pid = gp.get("PID") or gp.get("Pid") or gp.get("pid")
            if pid is None:
                continue
            f = flatten(gp, prefix="Gpu")
            f["GpuIndex"] = dev.get("Index", 0)
            f.setdefault("GpuTimestamp", pb.get("Timestamp", 0))
            by_pid[pid] = f
    return by_pid


def parse_jsonl(path, uuid):
    lines = []
    for i, raw in enumerate(open(path, "rb")):
        raw = raw.strip()
        if not raw:
            continue
        try:
            lines.append(json.loads(raw))
        except json.JSONDecodeError:
            print(f"  {uuid}: corrupt line {i}")

    if len(lines) < 2:
        return None, None, {}

    static = flatten(lines[0])
    metrics, procs = [], []

    for sample in lines[1:]:
        os_procs = sample.pop("Process", None) or []
        nv = sample.get("Nvidia") or []
        gp = gpu_procs(nv) if nv else {}
        for dev in nv:
            if isinstance(dev, dict) and "Processes" in dev:
                dev["Processes"].pop("List", None)

        metrics.append(flatten(sample))

        ts = sample.get("timestamp", 0)
        for proc in os_procs:
            pid = proc.get("Id") or proc.get("id")
            gpu_data = gp.get(pid) if pid is not None else None
            if (
                    gpu_data is None
                    and not proc.get("ResidentSetSize")
                    and not proc.get("VirtualMemoryBytes")
            ):
                continue
            fp = flatten(proc)
            t_vals = [
                v
                for k, v in fp.items()
                if k.endswith("T") and isinstance(v, (int, float)) and v > 0
            ]
            fp["timestamp"] = min(t_vals) if t_vals else ts
            if gpu_data:
                fp.update(gpu_data)
            procs.append(fp)

    mdf = strip_v(pd.DataFrame(metrics))
    mdf["uuid"], mdf["host"] = uuid, path.stem
    pdf = pd.DataFrame(procs) if procs else None
    if pdf is not None:
        pdf["uuid"], pdf["host"] = uuid, path.stem
    print(f"  {uuid}/{path.name}: {len(metrics)} samples, {len(procs)} procs")
    return mdf, pdf, static


def parse_results(path, uuid):
    if not path.exists():
        return {}, None, None
    data = json.loads(path.read_bytes())
    scalars = {k: v for k, v in data.items() if not isinstance(v, list)}
    arrays = {k: v for k, v in data.items() if isinstance(v, list)}
    itls = arrays.pop("itls", None)

    req_df = None
    if arrays:
        req_df = pd.DataFrame(arrays)
        req_df["request_index"] = req_df.index
        req_df["uuid"] = uuid

    tok_df = None
    if itls:
        tok_df = pd.DataFrame(
            {"request_index": range(len(itls)), "uuid": uuid, "itls": itls}
        )
        tok_df = tok_df.explode("itls").dropna(subset=["itls"])
        tok_df["itls"] = pd.to_numeric(tok_df["itls"])
        tok_df["itl_ms"] = tok_df["itls"] * 1000
        tok_df["token_position"] = tok_df.groupby("request_index").cumcount()
        tok_df = tok_df[["uuid", "request_index", "token_position", "itl_ms"]]

    return scalars, req_df, tok_df


def json_hist_to_arrays(df):
    hist_cols = [
        c
        for c in df.columns
        if c.endswith("Hist")
           and not c.endswith("HistT")
           and (df[c].dtype == object or pd.api.types.is_string_dtype(df[c]))
    ]
    if not hist_cols:
        return df

    for col in hist_cols:
        keys = None
        for val in reversed(df[col].values):
            if not isinstance(val, str):
                continue
            try:
                d = json.loads(val)
                if d:
                    finite = sorted((k for k in d if k != "+Inf"), key=float)
                    keys = finite + (["+Inf"] if "+Inf" in d else [])
                    break
            except (json.JSONDecodeError, TypeError):
                continue

        if not keys:
            df = df.drop(columns=[col])
            continue

        arrays = []
        for val in df[col]:
            if isinstance(val, str):
                try:
                    d = json.loads(val)
                    arrays.append([d.get(k, 0) for k in keys])
                    continue
                except (json.JSONDecodeError, TypeError):
                    pass
            arrays.append([0] * len(keys))
        df[col] = arrays

    return df


def main():
    args = sys.argv[1:]
    out_dir, roots, i = None, [], 0
    while i < len(args):
        if args[i] == "--out" and i + 1 < len(args):
            out_dir = Path(args[i + 1])
            i += 2
        else:
            roots.append(Path(args[i]))
            i += 1

    if not roots:
        print("Usage: ./collect.py [--out DIR] DIR [DIR ...]", file=sys.stderr)
        sys.exit(1)

    dirs = []
    for root in roots:
        if not root.is_dir():
            print(f"ERROR: {root} is not a directory", file=sys.stderr)
            sys.exit(1)
        found = sorted(
            d for d in root.iterdir() if d.is_dir() and (d / "run_log.jsonl").exists()
        )
        if not found:
            print(f"WARN: no subdirs with run_log.jsonl in {root}", file=sys.stderr)
        dirs.extend(found)

    if not dirs:
        print("ERROR: no experiment dirs found", file=sys.stderr)
        sys.exit(1)

    if out_dir is None:
        out_dir = roots[0]
    out_dir.mkdir(parents=True, exist_ok=True)
    print(f"Experiments: {[d.name for d in dirs]}\nOutput: {out_dir}/")

    all_runs, all_m, all_p, all_r, all_t = [], [], [], [], []

    for exp_dir in dirs:
        ds = exp_dir.name
        runs = [
            json.loads(ln) for ln in open(exp_dir / "run_log.jsonl", "rb") if ln.strip()
        ]
        print(f"\n[{ds}] {len(runs)} run(s)")

        for entry in runs:
            uid = entry["uuid"]
            subdir = exp_dir / uid
            if not subdir.is_dir():
                print(f"  {uid}: missing, skipping")
                continue

            scalars, req_df, tok_df = parse_results(subdir / "results.json", uid)
            if req_df is not None:
                req_df["dataset"] = ds
                all_r.append(req_df)
            if tok_df is not None:
                tok_df["dataset"] = ds
                all_t.append(tok_df)

            static_merged, host = {}, ""
            for jp in sorted(subdir.glob("*.jsonl")):
                mdf, pdf, static = parse_jsonl(jp, uid)
                if mdf is not None:
                    mdf["dataset"] = ds
                    all_m.append(mdf)
                    if static:
                        static_merged.update(static)
                    host = mdf["host"].iloc[0]
                if pdf is not None:
                    pdf["dataset"] = ds
                    all_p.append(pdf)

            all_runs.append(
                {
                    "uuid": uid,
                    "dataset": ds,
                    **{k: v for k, v in entry.items() if k != "uuid"},
                    **scalars,
                    **static_merged,
                    "host": host,
                }
            )

    print()
    if all_runs:
        df = pd.DataFrame(all_runs)
        for c in df.columns[df.dtypes == "object"]:
            try:
                df[c] = df[c].astype(float)
            except (ValueError, TypeError):
                pass
        df.to_parquet(out_dir / "runs.raw.parquet", index=False)
        print(f"runs.raw: {len(df)} rows, {len(df.columns)} cols")

    for name, frames in [
        ("procs", all_p),
        ("requests.raw", all_r),
        ("tokens.raw", all_t),
    ]:
        if frames:
            df = pd.concat(frames, ignore_index=True)
            df.to_parquet(out_dir / f"{name}.parquet", index=False)
            print(f"{name}: {len(df)} rows, {len(df.columns)} cols")

    if all_m:
        df = json_hist_to_arrays(pd.concat(all_m, ignore_index=True))
        df.to_parquet(out_dir / "metrics.raw.parquet", index=False)
        print(f"metrics.raw: {len(df)} rows, {len(df.columns)} cols")

    print("\nDone.")


if __name__ == "__main__":
    main()