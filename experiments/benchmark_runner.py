import os
import csv
import json
import tracemalloc
from typing import List, Dict
import logging

from solvers.cp_solver import solve_with_cpsat
from solvers.scip_solver import solve_with_scip
from experiments.generator import generate_benchmark_suite

logging.basicConfig(level=logging.INFO, format="%(message)s")
logger = logging.getLogger("benchmark")

def run_benchmarks() -> List[Dict]:
    suite = generate_benchmark_suite()
    results = []
    
    os.makedirs("experiments/results", exist_ok=True)
    csv_path = "experiments/results/benchmark_results.csv"
    json_path = "experiments/results/benchmark_results.json"

    logger.info("==================================================================")
    logger.info("   ROZPOCZĘCIE SERII POMIAROWEJ (CP-SAT vs SCIP MILP)")
    logger.info("==================================================================")
    logger.info(f"{'Instancja':<20} | {'Kat.':<6} | {'Zmienne':<7} | {'Solver':<7} | {'Status':<9} | {'Czas [s]':<8} | {'Pamięć [MB]':<11} | {'Obj Value'}")
    logger.info("-" * 95)

    for item in suite:
        name = item["id"]
        category = item["category"]
        req = item["request"]
        num_variables = req.num_employees * req.num_days * req.num_shifts

        for solver_type in ["cp-sat", "scip"]:
            tracemalloc.start()
            
            try:
                if solver_type == "cp-sat":
                    sol_res = solve_with_cpsat(req)
                else:
                    sol_res = solve_with_scip(req)
                    
                _, peak_memory = tracemalloc.get_traced_memory()
                peak_mb = peak_memory / (1024 * 1024)
            except Exception as e:
                peak_mb = 0.0
                sol_res = type("Obj", (), {
                    "solver_name": solver_type.upper(),
                    "status": f"ERROR: {str(e)[:15]}",
                    "solve_time_seconds": -1.0,
                    "objective_value": None
                })
            finally:
                tracemalloc.stop()

            record = {
                "instance_id": name,
                "category": category,
                "num_employees": req.num_employees,
                "num_days": req.num_days,
                "num_shifts": req.num_shifts,
                "num_variables": num_variables,
                "solver": sol_res.solver_name,
                "status": sol_res.status,
                "solve_time_seconds": round(sol_res.solve_time_seconds, 4),
                "peak_memory_mb": round(peak_mb, 2),
                "objective_value": sol_res.objective_value
            }
            results.append(record)

            obj_str = f"{sol_res.objective_value:.1f}" if sol_res.objective_value is not None else "N/A"
            logger.info(
                f"{name:<20} | {category:<6} | {num_variables:<7} | {sol_res.solver_name:<7} | "
                f"{sol_res.status:<9} | {sol_res.solve_time_seconds:<8.4f} | {peak_mb:<11.2f} | {obj_str}"
            )

    # Zapis do CSV
    keys = results[0].keys()
    with open(csv_path, mode="w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=keys)
        writer.writeheader()
        writer.writerows(results)

    # Zapis do JSON
    with open(json_path, mode="w", encoding="utf-8") as f:
        json.dump(results, f, indent=2, ensure_ascii=False)

    logger.info("-" * 95)
    logger.info(f"Pomiary zapisane w:\n - {csv_path}\n - {json_path}")
    return results

if __name__ == "__main__":
    run_benchmarks()
