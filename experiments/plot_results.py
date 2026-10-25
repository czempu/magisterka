import os
import json
import matplotlib.pyplot as plt
import numpy as np

def generate_charts():
    json_path = "experiments/results/benchmark_results.json"
    if not os.path.exists(json_path):
        print(f"Brak pliku wyników {json_path}. Najpierw uruchom benchmark_runner.py")
        return

    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    os.makedirs("experiments/figures", exist_ok=True)

    # Grupowanie danych per instancja i solver
    instances = []
    seen = set()
    for row in data:
        if row["instance_id"] not in seen:
            instances.append(row["instance_id"])
            seen.add(row["instance_id"])

    cp_times = []
    scip_times = []
    cp_mems = []
    scip_mems = []
    cp_objs = []
    scip_objs = []

    for inst in instances:
        cp_row = next((r for r in data if r["instance_id"] == inst and r["solver"] == "CP-SAT"), None)
        scip_row = next((r for r in data if r["instance_id"] == inst and r["solver"] == "SCIP"), None)

        cp_times.append(cp_row["solve_time_seconds"] if cp_row else 0)
        scip_times.append(scip_row["solve_time_seconds"] if scip_row else 0)

        cp_mems.append(cp_row["peak_memory_mb"] if cp_row else 0)
        scip_mems.append(scip_row["peak_memory_mb"] if scip_row else 0)

        cp_objs.append(cp_row["objective_value"] if cp_row and cp_row["objective_value"] is not None else 0)
        scip_objs.append(scip_row["objective_value"] if scip_row and scip_row["objective_value"] is not None else 0)

    x = np.arange(len(instances))
    width = 0.35

    # 1. WYKRES: Czas obliczeń (Solve Time) - skala logarytmiczna
    plt.figure(figsize=(10, 6))
    plt.bar(x - width/2, cp_times, width, label='CP-SAT (Google OR-Tools)', color='#1f77b4')
    plt.bar(x + width/2, scip_times, width, label='SCIP (PySCIPOpt MILP)', color='#ff7f0e')
    plt.yscale('log')
    plt.ylabel('Czas rozwiązywania [s] (skala log)')
    plt.title('Porównanie czasu rozwiązywania problemu harmonogramowania (CP-SAT vs SCIP)')
    plt.xticks(x, instances, rotation=25, ha='right')
    plt.legend()
    plt.grid(True, which="both", ls="--", alpha=0.5)
    plt.tight_layout()
    plt.savefig("experiments/figures/solve_time_comparison.png", dpi=300)
    plt.close()

    # 2. WYKRES: Zużycie pamięci (RAM Peak)
    plt.figure(figsize=(10, 6))
    plt.bar(x - width/2, cp_mems, width, label='CP-SAT', color='#2ca02c')
    plt.bar(x + width/2, scip_mems, width, label='SCIP', color='#d62728')
    plt.ylabel('Szczytowe zużycie pamięci RAM [MB]')
    plt.title('Porównanie zapotrzebowania na pamięć RAM (CP-SAT vs SCIP)')
    plt.xticks(x, instances, rotation=25, ha='right')
    plt.legend()
    plt.grid(True, ls="--", alpha=0.5)
    plt.tight_layout()
    plt.savefig("experiments/figures/memory_usage_comparison.png", dpi=300)
    plt.close()

    # 3. WYKRES: Wartość funkcji celu (Naruszenia preferencji)
    plt.figure(figsize=(10, 6))
    plt.plot(instances, cp_objs, marker='o', linewidth=2, label='CP-SAT (Obj Value)', color='#1f77b4')
    plt.plot(instances, scip_objs, marker='s', linewidth=2, linestyle='--', label='SCIP (Obj Value)', color='#ff7f0e')
    plt.ylabel('Wartość funkcji celu (Suma kar za preferencje)')
    plt.title('Jakość znalezionych rozwiązań w funkcji wielkości instancji')
    plt.xticks(rotation=25, ha='right')
    plt.legend()
    plt.grid(True, ls="--", alpha=0.5)
    plt.tight_layout()
    plt.savefig("experiments/figures/objective_value_comparison.png", dpi=300)
    plt.close()

    print("Wykresy zostały pomyślnie wygenerowane w 'experiments/figures/':")
    print(" - solve_time_comparison.png")
    print(" - memory_usage_comparison.png")
    print(" - objective_value_comparison.png")

if __name__ == "__main__":
    generate_charts()
